package tfconf

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hrmsk66/terraformify/pkg/cli"
	"github.com/hrmsk66/terraformify/pkg/file"
	"github.com/hrmsk66/terraformify/pkg/naming"
	"github.com/hrmsk66/terraformify/pkg/prop"
	"github.com/hrmsk66/terraformify/pkg/tfstate"
	"github.com/zclconf/go-cty/cty"
)

var (
	ErrAttrNotFound = errors.New("attribute not found")
)

type TFConf struct {
	*hclwrite.File
}

type SensitiveAttr struct {
	BlockType string
	Key       string
	Value     string
}

func Load(rawHCL string) (*TFConf, error) {
	// Clean up rawHCL to prevent parser errors from occurring
	t := cleanupHCL(rawHCL)

	f, diags := hclwrite.ParseConfig([]byte(t), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("errors: %s", diags)
	}

	return &TFConf{f}, nil
}

func (tfconf *TFConf) ParseServiceResource(serviceProp prop.TFBlock, c *cli.Config) ([]prop.TFBlock, error) {
	// Check top-level blocks
	for _, block := range tfconf.Body().Blocks() {
		id, err := getStringAttributeValue(block, "id")
		if err != nil {
			return nil, err
		}

		if t := block.Type(); t != "resource" {
			return nil, fmt.Errorf("unexpected block type: %v", t)
		}

		if id != c.ID {
			log.Printf("[INFO] tfconf: skip %s (%s)", block.Labels()[0], id)
			continue
		}

		log.Printf("[INFO] tfconf: found %s (%s)", block.Labels()[0], id)
		body := block.Body()

		// Get the nested blocks
		nestedBlocks := body.Blocks()

		// Collect block properties that require surgical changes.
		props := make([]prop.TFBlock, 0, len(nestedBlocks))

		for _, block := range nestedBlocks {
			blockType := block.Type()

			switch blockType {
			case "acl":
				id, err := getStringAttributeValue(block, "acl_id")
				if err != nil {
					return nil, err
				}
				name, err := getStringAttributeValue(block, "name")
				if err != nil {
					return nil, err
				}
				prop := prop.NewACLResource(id, name, serviceProp)
				props = append(props, prop)
			case "dictionary":
				id, err := getStringAttributeValue(block, "dictionary_id")
				if err != nil {
					return nil, err
				}
				name, err := getStringAttributeValue(block, "name")
				if err != nil {
					return nil, err
				}
				prop := prop.NewDictionaryResource(id, name, serviceProp)
				props = append(props, prop)
			case "waf":
				id, err := getStringAttributeValue(block, "waf_id")
				if err != nil {
					return nil, err
				}
				c.WafID = id
				prop := prop.NewWAFResource(id, serviceProp)
				props = append(props, prop)
			case "dynamicsnippet":
				id, err := getStringAttributeValue(block, "snippet_id")
				if err != nil {
					return nil, err
				}
				name, err := getStringAttributeValue(block, "name")
				if err != nil {
					return nil, err
				}
				prop := prop.NewDynamicSnippetResource(id, name, serviceProp)
				props = append(props, prop)
			case "resource_link":
				id, err := getStringAttributeValue(block, "resource_id")
				if err != nil {
					return nil, err
				}
				name, err := getStringAttributeValue(block, "name")
				if err != nil {
					return nil, err
				}
				prop := prop.NewLinkedResource(id, name, serviceProp)
				props = append(props, prop)
			}
		}
		return props, nil
	}
	return nil, errors.New("tfconf: target service resource not found")
}

func (tfconf *TFConf) RewriteResources(serviceProp prop.TFBlock, props []prop.TFBlock, c *cli.Config) ([]SensitiveAttr, error) {
	// Read terraform.tfstate into the variable
	state, err := tfstate.Load(c.Directory)
	if err != nil {
		return nil, err
	}

	var sensitiveAttrs []SensitiveAttr
	// Read resource blocks
	for _, block := range tfconf.Body().Blocks() {
		if t := block.Type(); t != "resource" {
			return nil, fmt.Errorf("unexpected block type: %v", t)
		}

		var id string
		switch block.Labels()[0] {
		case "fastly_service_vcl":
			id, err = getStringAttributeValue(block, "id")
			if err != nil {
				return nil, err
			}
			if id != c.ID {
				tfconf.Body().RemoveBlock(block)
				continue
			}

			sensitiveAttrs, err = rewriteVCLServiceResource(block, serviceProp, state, c)
			if err != nil {
				return nil, err
			}
		case "fastly_service_compute":
			id, err = getStringAttributeValue(block, "id")
			if err != nil {
				return nil, err
			}
			if id != c.ID {
				tfconf.Body().RemoveBlock(block)
				continue
			}

			// Add "fastly_package_hash" data block
			appendFastlyPackageHashBlock(tfconf, serviceProp, c)

			sensitiveAttrs, err = rewriteComputeServiceResource(block, serviceProp, props, state, c)
			if err != nil {
				return nil, err
			}
		case "fastly_configstore", "fastly_secretstore", "fastly_kvstore":
			rewriteLinkedResource(block)
		case "fastly_configstore_entries":
			err = rewriteConfigStoreEntries(block, props, c)
			if err != nil {
				return nil, err
			}
		case "fastly_service_waf_configuration":
			id, err := getStringAttributeValue(block, "waf_id")
			if err != nil {
				return nil, err
			}
			if id != c.WafID {
				tfconf.Body().RemoveBlock(block)
				continue
			}

			err = rewriteWAFResource(block, serviceProp)
			if err != nil {
				return nil, err
			}
		case "fastly_service_dynamic_snippet_content":
			sid, err := getStringAttributeValue(block, "service_id")
			if err != nil {
				return nil, err
			}
			if sid != c.ID {
				tfconf.Body().RemoveBlock(block)
				continue
			}

			err = rewriteDynamicSnippetResource(block, serviceProp, state, c)
			if err != nil {
				return nil, err
			}
		case "fastly_service_dictionary_items":
			sid, err := getStringAttributeValue(block, "service_id")
			if err != nil {
				return nil, err
			}
			if sid != c.ID {
				tfconf.Body().RemoveBlock(block)
				continue
			}

			err = rewriteDictionaryResource(block, serviceProp, state, c)
			if err != nil {
				return nil, err
			}
		case "fastly_service_acl_entries":
			sid, err := getStringAttributeValue(block, "service_id")
			if err != nil {
				return nil, err
			}
			if sid != c.ID {
				tfconf.Body().RemoveBlock(block)
				continue
			}

			err = rewriteACLResource(block, serviceProp, state, c)
			if err != nil {
				return nil, err
			}
		// Skip handling unknown resource blocks
		default:
			tfconf.Body().RemoveBlock(block)
			continue
		}
	}

	return sensitiveAttrs, nil
}

func rewriteVCLServiceResource(block *hclwrite.Block, serviceProp prop.TFBlock, s *tfstate.TFState, c *cli.Config) ([]SensitiveAttr, error) {
	var sensitiveAttrs []SensitiveAttr

	st, err := s.AddTemplate(tfstate.ServiceQueryTmplate)
	if err != nil {
		return nil, err
	}

	// Remove read-only attributes
	body := block.Body()
	body.RemoveAttribute("id")
	body.RemoveAttribute("active_version")
	body.RemoveAttribute("cloned_version")
	body.RemoveAttribute("imported")
	body.RemoveAttribute("force_refresh")

	// If no service level comments are set, set blank
	// Otherwise, Terraform will set `Managed by Terraform` and cause a configuration diff
	comment, err := getStringAttributeValue(block, "comment")
	if err != nil {
		if !errors.Is(err, ErrAttrNotFound) {
			return nil, err
		}

		if comment == "" {
			body.SetAttributeValue("comment", cty.StringVal(""))
		}
	}

	if c.ForceDestroy {
		body.AppendNewline()
		body.SetAttributeValue("force_destroy", cty.BoolVal(true))
	}

	for _, block := range body.Blocks() {
		blockType := block.Type()
		nestedBody := block.Body()

		switch blockType {
		case "acl":
			nestedBody.RemoveAttribute("acl_id")
			if c.ForceDestroy {
				nestedBody.SetAttributeValue("force_destroy", cty.BoolVal(true))
			}
		case "dictionary":
			nestedBody.RemoveAttribute("dictionary_id")
			if c.ForceDestroy {
				nestedBody.SetAttributeValue("force_destroy", cty.BoolVal(true))
			}
		case "waf":
			nestedBody.RemoveAttribute("waf_id")
		case "dynamicsnippet":
			nestedBody.RemoveAttribute("snippet_id")
		case "product_enablement":
			nestedBody.RemoveAttribute("name")
		case "request_setting":
			nestedBody.RemoveAttribute("geo_headers")

			// Get name from TFConf
			name, err := getStringAttributeValue(block, "name")
			if err != nil {
				return nil, err
			}

			// Get content from TFState
			v, err := st.ServiceQuery(tfstate.ServiceQueryParams{
				ServiceId:       c.ID,
				NestedBlockName: blockType,
				Name:            name,
				AttributeName:   "xff",
			})
			if err != nil {
				return nil, err
			}

			// In the provider schema, xff is an optional attribute with a default value of "append"
			// Because of the default value, Terraform attempts to add the default value even if the value is not set for the actual service.
			// To workaround the issue, explicitly setting xff attribute with blank value if it's blank in the state file
			if v.String() == "" {
				nestedBody.SetAttributeValue("xff", cty.StringVal(""))
			}
		case "response_object":
			// Get name from TFConf
			name, err := getStringAttributeValue(block, "name")
			if err != nil {
				return nil, err
			}

			// Get content from TFState
			v, err := st.ServiceQuery(tfstate.ServiceQueryParams{
				ServiceId:       c.ID,
				NestedBlockName: blockType,
				Name:            name,
				AttributeName:   "content",
			})
			if err != nil {
				return nil, err
			}

			ext := "txt"
			filename := fmt.Sprintf("%s.%s", naming.Normalize(name), ext)
			if err = file.WriteContent(c.Directory, c.ResourceName, filename, v.Bytes()); err != nil {
				return nil, err
			}

			// Replace content attribute of the nested block with file function expression
			path := filepath.Join(".", "content", c.ResourceName, filename)
			tokens := buildFileFunction(path)
			nestedBody.SetAttributeRaw("content", tokens)
		case "snippet":
			// Get name from TFConf
			name, err := getStringAttributeValue(block, "name")
			if err != nil {
				return nil, err
			}

			// Get content from TFState
			v, err := st.ServiceQuery(tfstate.ServiceQueryParams{
				ServiceId:       c.ID,
				NestedBlockName: blockType,
				Name:            name,
				AttributeName:   "content",
			})
			if err != nil {
				return nil, err
			}

			// Save content to a file
			filename := fmt.Sprintf("snippet_%s.vcl", naming.Normalize(name))
			if err = file.WriteVCL(c.Directory, c.ResourceName, filename, v.Bytes()); err != nil {
				return nil, err
			}

			// Replace content attribute of the nested block with file function expression
			path := filepath.Join(".", "vcl", c.ResourceName, filename)
			tokens := buildFileFunction(path)
			nestedBody.SetAttributeRaw("content", tokens)
		case "vcl":
			// Get name from TFConf
			name, err := getStringAttributeValue(block, "name")
			if err != nil {
				return nil, err
			}

			// Get content from TFState
			v, err := st.ServiceQuery(tfstate.ServiceQueryParams{
				ServiceId:       c.ID,
				NestedBlockName: blockType,
				Name:            name,
				AttributeName:   "content",
			})
			if err != nil {
				return nil, err
			}

			// Save content to a file
			filename := fmt.Sprintf("%s.vcl", naming.Normalize(name))
			if err = file.WriteVCL(c.Directory, c.ResourceName, filename, v.Bytes()); err != nil {
				return nil, err
			}

			// Replace content attribute of the nested block with file function expression
			path := filepath.Join(".", "vcl", c.ResourceName, filename)
			tokens := buildFileFunction(path)
			nestedBody.SetAttributeRaw("content", tokens)
		case "backend":
			name, err := getStringAttributeValue(block, "name")
			if err != nil {
				return nil, err
			}

			// Handling sensitive attrs
			keys := []string{"ssl_client_cert", "ssl_client_key"}
			for _, key := range keys {
				v, err := st.ServiceQuery(tfstate.ServiceQueryParams{
					ServiceId:       c.ID,
					NestedBlockName: blockType,
					Name:            name,
					AttributeName:   key,
				})
				if err != nil {
					return nil, err
				}
				if v.String() != "" {
					varName := naming.Normalize(name) + "_" + key
					nestedBody.SetAttributeTraversal(key, buildVariableRef(varName))
					sensitiveAttrs = append(sensitiveAttrs, SensitiveAttr{blockType, varName, v.String()})
				}
			}
		default:
			if strings.HasPrefix(blockType, "logging_") {
				name, err := getStringAttributeValue(block, "name")
				if err != nil {
					return nil, err
				}

				format, err := st.ServiceQuery(tfstate.ServiceQueryParams{
					ServiceId:       c.ID,
					NestedBlockName: blockType,
					Name:            name,
					AttributeName:   "format",
				})
				if err != nil {
					return nil, err
				}

				ext := "txt"
				if json.Valid(format.Bytes()) {
					ext = "json"
				}
				filename := fmt.Sprintf("%s.%s", naming.Normalize(name), ext)
				if err = file.WriteLogFormat(c.Directory, c.ResourceName, filename, format.Bytes()); err != nil {
					return nil, err
				}
				// Replace content attribute of the nested block with file function expression
				path := filepath.Join(".", "logformat", c.ResourceName, filename)
				tokens := buildFileFunction(path)
				nestedBody.SetAttributeRaw("format", tokens)

				// Handling sensitive attrs
				var keys []string
				switch blockType {
				case "logging_bigquery":
					keys = []string{"email", "secret_key"}
				case "logging_blobstorage":
					keys = []string{"sas_token"}
				case "logging_cloudfiles":
					keys = []string{"access_key"}
				case "logging_datadog":
					keys = []string{"token"}
				case "logging_digitalocean":
					keys = []string{"access_key", "secret_key"}
				case "logging_elasticsearch":
					keys = []string{"password", "tls_client_key"}
				case "logging_ftp":
					keys = []string{"password"}
				case "logging_gcs":
					keys = []string{"secret_key"}
				case "logging_googlepubsub":
					keys = []string{"secret_key"}
				case "logging_heroku":
					keys = []string{"token"}
				case "logging_honeycomb":
					keys = []string{"token"}
				case "logging_https":
					keys = []string{"tls_client_key"}
				case "logging_kafka":
					keys = []string{"password", "tls_client_key"}
				case "logging_kinesis":
					keys = []string{"access_key", "secret_key"}
				case "logging_loggly":
					keys = []string{"token"}
				case "logging_logshuttle":
					keys = []string{"token"}
				case "logging_newrelic":
					keys = []string{"token"}
				case "logging_openstack":
					keys = []string{"access_key"}
				case "logging_s3":
					// Need S3 keys when "s3_iam_role" is empty
					v, err := st.ServiceQuery(tfstate.ServiceQueryParams{
						ServiceId:       c.ID,
						NestedBlockName: blockType,
						Name:            name,
						AttributeName:   "s3_iam_role",
					})
					if err != nil {
						return nil, err
					}
					if v.String() == "" {
						keys = []string{"s3_access_key", "s3_secret_key"}
					}
				case "logging_scalyr":
					keys = []string{"token"}
				case "logging_sftp":
					keys = []string{"password", "secret_key"}
				case "logging_splunk":
					keys = []string{"tls_client_key", "token"}
				case "logging_syslog":
					keys = []string{"tls_client_key"}
				}
				for _, key := range keys {
					v, err := st.ServiceQuery(tfstate.ServiceQueryParams{
						ServiceId:       c.ID,
						NestedBlockName: blockType,
						Name:            name,
						AttributeName:   key,
					})
					if err != nil {
						return nil, err
					}

					// the attribute names for under "logging_s3" are redundant. Removing the prefix "s3_" in the variable names
					varName := naming.Normalize(name) + "_" + strings.TrimPrefix(key, "s3_")
					nestedBody.SetAttributeTraversal(key, buildVariableRef(varName))
					sensitiveAttrs = append(sensitiveAttrs, SensitiveAttr{blockType, varName, v.String()})
				}
			}
		}
	}

	return sensitiveAttrs, nil
}

func rewriteComputeServiceResource(block *hclwrite.Block, serviceProp prop.TFBlock, props []prop.TFBlock, s *tfstate.TFState, c *cli.Config) ([]SensitiveAttr, error) {
	var sensitiveAttrs []SensitiveAttr

	st, err := s.AddTemplate(tfstate.ServiceQueryTmplate)
	if err != nil {
		return nil, err
	}

	// Remove read-only attributes
	body := block.Body()
	body.RemoveAttribute("id")
	body.RemoveAttribute("active_version")
	body.RemoveAttribute("cloned_version")
	body.RemoveAttribute("imported")
	body.RemoveAttribute("force_refresh")

	// If no service level comments are set, set blank
	// Otherwise, Terraform will set `Managed by Terraform` and cause a configuration diff
	comment, err := getStringAttributeValue(block, "comment")
	if err != nil {
		if !errors.Is(err, ErrAttrNotFound) {
			return nil, err
		}

		if comment == "" {
			body.SetAttributeValue("comment", cty.StringVal(""))
		}
	}

	if c.ForceDestroy {
		body.AppendNewline()
		body.SetAttributeValue("force_destroy", cty.BoolVal(true))
	}

	for _, block := range body.Blocks() {
		blockType := block.Type()
		nestedBody := block.Body()

		switch blockType {
		case "dictionary":
			nestedBody.RemoveAttribute("dictionary_id")
			if c.ForceDestroy {
				nestedBody.SetAttributeValue("force_destroy", cty.BoolVal(true))
			}
		case "product_enablement":
			nestedBody.RemoveAttribute("name")
		case "package":
			nestedBody.SetAttributeTraversal("filename", buildPackageHashRef(serviceProp, "filename"))
			nestedBody.SetAttributeTraversal("source_code_hash", buildPackageHashRef(serviceProp, "hash"))
		case "resource_link":
			resourceId, err := getStringAttributeValue(block, "resource_id")
			if err != nil {
				return nil, err
			}
			for _, prop := range props {
				if prop.GetID() == resourceId {
					nestedBody.SetAttributeTraversal("name", buildResourceRef(prop, "name"))
					nestedBody.SetAttributeTraversal("resource_id", buildResourceRef(prop, "id"))
					break
				}
			}
			nestedBody.RemoveAttribute("link_id")
		case "backend":
			name, err := getStringAttributeValue(block, "name")
			if err != nil {
				return nil, err
			}

			// Handling sensitive attrs
			keys := []string{"ssl_client_cert", "ssl_client_key"}
			for _, key := range keys {
				v, err := st.ServiceQuery(tfstate.ServiceQueryParams{
					ServiceId:       c.ID,
					NestedBlockName: blockType,
					Name:            name,
					AttributeName:   key,
				})
				if err != nil {
					return nil, err
				}
				if v.String() != "" {
					varName := naming.Normalize(name) + "_" + key
					nestedBody.SetAttributeTraversal(key, buildVariableRef(varName))
					sensitiveAttrs = append(sensitiveAttrs, SensitiveAttr{blockType, varName, v.String()})
				}
			}
		default:
			if strings.HasPrefix(blockType, "logging_") {
				name, err := getStringAttributeValue(block, "name")
				if err != nil {
					return nil, err
				}

				// Handling sensitive attrs
				var keys []string
				switch blockType {
				case "logging_bigquery":
					keys = []string{"email", "secret_key"}
				case "logging_blobstorage":
					keys = []string{"sas_token"}
				case "logging_cloudfiles":
					keys = []string{"access_key"}
				case "logging_datadog":
					keys = []string{"token"}
				case "logging_digitalocean":
					keys = []string{"access_key", "secret_key"}
				case "logging_elasticsearch":
					keys = []string{"password", "tls_client_key"}
				case "logging_ftp":
					keys = []string{"password"}
				case "logging_gcs":
					keys = []string{"secret_key"}
				case "logging_googlepubsub":
					keys = []string{"secret_key"}
				case "logging_heroku":
					keys = []string{"token"}
				case "logging_honeycomb":
					keys = []string{"token"}
				case "logging_https":
					keys = []string{"tls_client_key"}
				case "logging_kafka":
					keys = []string{"password", "tls_client_key"}
				case "logging_kinesis":
					keys = []string{"access_key", "secret_key"}
				case "logging_loggly":
					keys = []string{"token"}
				case "logging_logshuttle":
					keys = []string{"token"}
				case "logging_newrelic":
					keys = []string{"token"}
				case "logging_openstack":
					keys = []string{"access_key"}
				case "logging_s3":
					keys = []string{"s3_access_key", "s3_secret_key"}
				case "logging_scalyr":
					keys = []string{"token"}
				case "logging_sftp":
					keys = []string{"password", "secret_key"}
				case "logging_splunk":
					keys = []string{"tls_client_key", "token"}
				case "logging_syslog":
					keys = []string{"tls_client_key"}
				}
				for _, key := range keys {
					v, err := st.ServiceQuery(tfstate.ServiceQueryParams{
						ServiceId:       c.ID,
						NestedBlockName: blockType,
						Name:            name,
						AttributeName:   key,
					})
					if err != nil {
						return nil, err
					}

					// the attribute names for under "logging_s3" are redundant. Removing the prefix "s3_" in the variable names
					varName := naming.Normalize(name) + "_" + strings.TrimPrefix(key, "s3_")
					nestedBody.SetAttributeTraversal(key, buildVariableRef(varName))
					sensitiveAttrs = append(sensitiveAttrs, SensitiveAttr{blockType, varName, v.String()})
				}
			}
		}
	}

	return sensitiveAttrs, nil
}

func rewriteLinkedResource(block *hclwrite.Block) {
	// remove read-only attributes
	body := block.Body()
	body.RemoveAttribute("id")
}

func rewriteConfigStoreEntries(block *hclwrite.Block, props []prop.TFBlock, c *cli.Config) error {
	// remove read-only attributes
	body := block.Body()
	body.RemoveAttribute("id")

	storeId, err := getStringAttributeValue(block, "store_id")
	if err != nil {
		return err
	}
	for _, prop := range props {
		if prop.GetID() == storeId {
			body.SetAttributeTraversal("store_id", buildResourceRef(prop, "id"))
			break
		}
	}

	if c.ManageAll {
		body.SetAttributeValue("manage_entries", cty.BoolVal(true))
	}

	return nil
}

func rewriteACLResource(block *hclwrite.Block, serviceProp prop.TFBlock, s *tfstate.TFState, c *cli.Config) error {
	if err := rewriteCommonAttributes(block, serviceProp, s, c); err != nil {
		return err
	}

	// remove read-only attributes from each ACL entry
	body := block.Body()
	for _, block := range body.Blocks() {
		t := block.Type()
		nb := block.Body()
		if t != "entry" {
			return fmt.Errorf("unexpected Terraform block: %#v", block)
		}
		nb.RemoveAttribute("id")
	}

	if c.ManageAll {
		body.SetAttributeValue("manage_entries", cty.BoolVal(true))
	}

	return nil
}

func rewriteDictionaryResource(block *hclwrite.Block, serviceProp prop.TFBlock, s *tfstate.TFState, c *cli.Config) error {
	if err := rewriteCommonAttributes(block, serviceProp, s, c); err != nil {
		return err
	}

	body := block.Body()
	if c.ManageAll {
		body.SetAttributeValue("manage_items", cty.BoolVal(true))
	}

	return nil
}

func rewriteDynamicSnippetResource(block *hclwrite.Block, serviceProp prop.TFBlock, s *tfstate.TFState, c *cli.Config) error {
	if err := rewriteCommonAttributes(block, serviceProp, s, c); err != nil {
		return err
	}

	// replace content value with file()
	name := block.Labels()[1]

	// Get content from the state file
	st, err := s.AddTemplate(tfstate.DsnippetQueryTmplate)
	if err != nil {
		return err
	}
	v, err := st.DSnippetQuery(tfstate.DSnippetQueryParams{
		ResourceName: name,
	})
	if err != nil {
		return err
	}

	// Save content to a file
	filename := fmt.Sprintf("dsnippet_%s.vcl", naming.Normalize(name))
	if err = file.WriteVCL(c.Directory, c.ResourceName, filename, v.Bytes()); err != nil {
		return err
	}

	// Replace content attribute with file function expression
	body := block.Body()
	path := filepath.Join(".", "vcl", c.ResourceName, filename)
	tokens := buildFileFunction(path)
	body.SetAttributeRaw("content", tokens)

	if c.ManageAll {
		body.SetAttributeValue("manage_snippets", cty.BoolVal(true))
	}

	return nil
}

func rewriteCommonAttributes(block *hclwrite.Block, serviceProp prop.TFBlock, s *tfstate.TFState, c *cli.Config) error {
	var idName, attrName string
	switch block.Labels()[0] {
	case "fastly_service_dynamic_snippet_content":
		idName = "snippet_id"
		attrName = "dynamicsnippet"
	case "fastly_service_dictionary_items":
		idName = "dictionary_id"
		attrName = "dictionary"
	case "fastly_service_acl_entries":
		idName = "acl_id"
		attrName = "acl"
	}

	// Getting the name of the resource from the state file
	id, err := getStringAttributeValue(block, idName)
	if err != nil {
		return err
	}
	st, err := s.AddTemplate(tfstate.ResourceNameQueryTmplate)
	if err != nil {
		return err
	}
	name, err := st.ResourceNameQuery(tfstate.ResourceNameQueryParams{
		ResourceType:    serviceProp.GetType(),
		NestedBlockName: attrName,
		IDName:          idName,
		ID:              id,
	})
	if err != nil {
		return err
	}

	body := block.Body()

	// Add for_each to the resource block
	body.AppendNewline()
	tokens := buildForEach(serviceProp, attrName, name.String())
	body.SetAttributeRaw("for_each", tokens)

	// Setting the resource ID (acl_id, dictionary_id, snippet_id)
	resourceIDRef := buildForEachIDRef(idName)
	body.SetAttributeTraversal(idName, resourceIDRef)

	// remove read-only attributes
	body.RemoveAttribute("id")

	// set service_id to represent the resource dependency
	ref := buildServiceIDRef(serviceProp)
	body.SetAttributeTraversal("service_id", ref)

	return nil
}

func rewriteWAFResource(block *hclwrite.Block, serviceProp prop.TFBlock) error {
	body := block.Body()
	// remove read-only attributes
	body.RemoveAttribute("active")
	body.RemoveAttribute("cloned_version")
	body.RemoveAttribute("number")
	body.RemoveAttribute("id")

	// set waf_id to represent the resource dependency
	body.SetAttributeTraversal("waf_id", hcl.Traversal{
		hcl.TraverseRoot{Name: serviceProp.GetType()},
		hcl.TraverseAttr{Name: serviceProp.GetNormalizedName()},
		hcl.TraverseAttr{Name: "waf"},
		hcl.TraverseIndex{Key: cty.NumberUIntVal(0)},
		hcl.TraverseAttr{Name: "waf_id"},
	})

	return nil
}

func appendFastlyPackageHashBlock(tfconf *TFConf, serviceProp prop.TFBlock, config *cli.Config) {
	tfconf.Body().AppendNewline()
	p := tfconf.Body().AppendNewBlock("data", []string{"fastly_package_hash", serviceProp.GetNormalizedName()})
	p.Body().SetAttributeValue("filename", cty.StringVal(config.Package))
}

func getStringAttributeValue(block *hclwrite.Block, attrKey string) (string, error) {
	// find TokenQuotedLit
	attr := block.Body().GetAttribute(attrKey)
	if attr == nil {
		return "", fmt.Errorf(`%w: failed to find "%s" in "%s"`, ErrAttrNotFound, attrKey, block.Type())
	}

	expr := attr.Expr()
	exprTokens := expr.BuildTokens(nil)

	i := 0
	for i < len(exprTokens) && exprTokens[i].Type != hclsyntax.TokenQuotedLit {
		i++
	}

	if i == len(exprTokens) {
		return "", fmt.Errorf("failed to find TokenQuotedLit: %#v", attr)
	}

	value := string(exprTokens[i].Bytes)
	return value, nil
}

func buildFilesha512Function(path string) hclwrite.Tokens {
	return hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("filesha512")},
		{Type: hclsyntax.TokenOParen, Bytes: []byte{'('}},
		{Type: hclsyntax.TokenOQuote, Bytes: []byte{'"'}},
		{Type: hclsyntax.TokenQuotedLit, Bytes: []byte(path)},
		{Type: hclsyntax.TokenCQuote, Bytes: []byte{'"'}},
		{Type: hclsyntax.TokenCParen, Bytes: []byte{')'}},
	}
}

func buildFileFunction(path string) hclwrite.Tokens {
	return hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("file")},
		{Type: hclsyntax.TokenOParen, Bytes: []byte{'('}},
		{Type: hclsyntax.TokenOQuote, Bytes: []byte{'"'}},
		{Type: hclsyntax.TokenQuotedLit, Bytes: []byte(path)},
		{Type: hclsyntax.TokenCQuote, Bytes: []byte{'"'}},
		{Type: hclsyntax.TokenCParen, Bytes: []byte{')'}},
	}
}

func buildForEach(serviceProp prop.TFBlock, resourceType, name string) hclwrite.Tokens {
	return hclwrite.Tokens{
		{Type: hclsyntax.TokenOBrace, Bytes: []byte{'{'}, SpacesBefore: 1},
		{Type: hclsyntax.TokenNewline, Bytes: []byte("\n"), SpacesBefore: 0},
		{Type: hclsyntax.TokenIdent, Bytes: []byte("for"), SpacesBefore: 2},
		{Type: hclsyntax.TokenIdent, Bytes: []byte{resourceType[0]}, SpacesBefore: 1},
		{Type: hclsyntax.TokenIdent, Bytes: []byte("in"), SpacesBefore: 1},
		{Type: hclsyntax.TokenIdent, Bytes: []byte(serviceProp.GetType()), SpacesBefore: 1},
		{Type: hclsyntax.TokenDot, Bytes: []byte{'.'}, SpacesBefore: 0},
		{Type: hclsyntax.TokenIdent, Bytes: []byte(serviceProp.GetNormalizedName()), SpacesBefore: 0},
		{Type: hclsyntax.TokenDot, Bytes: []byte{'.'}, SpacesBefore: 0},
		{Type: hclsyntax.TokenIdent, Bytes: []byte(resourceType), SpacesBefore: 0},
		{Type: hclsyntax.TokenColon, Bytes: []byte{':'}, SpacesBefore: 1},
		{Type: hclsyntax.TokenIdent, Bytes: []byte{resourceType[0]}, SpacesBefore: 1},
		{Type: hclsyntax.TokenDot, Bytes: []byte{'.'}, SpacesBefore: 0},
		{Type: hclsyntax.TokenIdent, Bytes: []byte("name"), SpacesBefore: 0},
		{Type: hclsyntax.TokenFatArrow, Bytes: []byte("=>"), SpacesBefore: 1},
		{Type: hclsyntax.TokenIdent, Bytes: []byte{resourceType[0]}, SpacesBefore: 1},
		{Type: hclsyntax.TokenIdent, Bytes: []byte("if"), SpacesBefore: 1},
		{Type: hclsyntax.TokenIdent, Bytes: []byte{resourceType[0]}, SpacesBefore: 1},
		{Type: hclsyntax.TokenDot, Bytes: []byte{'.'}, SpacesBefore: 0},
		{Type: hclsyntax.TokenIdent, Bytes: []byte("name"), SpacesBefore: 0},
		{Type: hclsyntax.TokenEqualOp, Bytes: []byte("=="), SpacesBefore: 1},
		{Type: hclsyntax.TokenOQuote, Bytes: []byte{'"'}, SpacesBefore: 1},
		{Type: hclsyntax.TokenQuotedLit, Bytes: []byte(name), SpacesBefore: 0},
		{Type: hclsyntax.TokenCQuote, Bytes: []byte{'"'}, SpacesBefore: 0},
		{Type: hclsyntax.TokenNewline, Bytes: []byte("\n"), SpacesBefore: 0},
		{Type: hclsyntax.TokenCBrace, Bytes: []byte{'}'}, SpacesBefore: 2},
	}
}

func buildForEachIDRef(idName string) hcl.Traversal {
	return hcl.Traversal{
		hcl.TraverseRoot{Name: "each"},
		hcl.TraverseAttr{Name: "value"},
		hcl.TraverseAttr{Name: idName},
	}
}

func buildServiceIDRef(serviceProp prop.TFBlock) hcl.Traversal {
	return hcl.Traversal{
		hcl.TraverseRoot{Name: serviceProp.GetType()},
		hcl.TraverseAttr{Name: serviceProp.GetNormalizedName()},
		hcl.TraverseAttr{Name: "id"},
	}
}

func buildVariableRef(varName string) hcl.Traversal {
	return hcl.Traversal{
		hcl.TraverseRoot{Name: "var"},
		hcl.TraverseAttr{Name: varName},
	}
}

func buildPackageHashRef(prop prop.TFBlock, attr string) hcl.Traversal {
	return hcl.Traversal{
		hcl.TraverseRoot{Name: "data"},
		hcl.TraverseAttr{Name: "fastly_package_hash"},
		hcl.TraverseAttr{Name: prop.GetNormalizedName()},
		hcl.TraverseAttr{Name: attr},
	}
}

func buildResourceRef(prop prop.TFBlock, attr string) hcl.Traversal {
	return hcl.Traversal{
		hcl.TraverseRoot{Name: prop.GetType()},
		hcl.TraverseAttr{Name: prop.GetNormalizedName()},
		hcl.TraverseAttr{Name: attr},
	}
}

func BuildTFVars(attrs []SensitiveAttr) []byte {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	for i, attr := range attrs {
		if i != 0 {
			rootBody.AppendNewline()
		}

		rootBody.SetAttributeValue(attr.Key, cty.StringVal(attr.Value))
	}

	return f.Bytes()
}

func BuildVariableDefinitions(attrs []SensitiveAttr) []byte {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	for i, attr := range attrs {
		if i != 0 {
			rootBody.AppendNewline()
		}

		varBlock := rootBody.AppendNewBlock("variable", []string{attr.Key})
		varBody := varBlock.Body()

		varBody.SetAttributeValue("description", cty.StringVal(strings.ReplaceAll(attr.Key, "_", " ")))
		varBody.SetAttributeRaw("type", hclwrite.Tokens{{Type: hclsyntax.TokenIdent, Bytes: []byte("string")}})
		varBody.SetAttributeValue("sensitive", cty.BoolVal(true))
	}

	return f.Bytes()
}
