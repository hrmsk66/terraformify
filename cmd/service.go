package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hrmsk66/terraformify/pkg/cli"
	"github.com/hrmsk66/terraformify/pkg/file"
	"github.com/hrmsk66/terraformify/pkg/prop"
	"github.com/hrmsk66/terraformify/pkg/terraform"
	"github.com/hrmsk66/terraformify/pkg/tfconf"
	"github.com/hrmsk66/terraformify/pkg/tfstate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:          "service <service-id>",
	Short:        "Generate TF files for an existing Fastly service",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := cli.CreateLogFilter()
		log.SetOutput(filter)
		log.Printf("[INFO] CLI version: %s", version)

		workingDir, err := cmd.Flags().GetString("working-dir")
		if err != nil {
			return err
		}
		err = cli.CheckDirEmpty(workingDir)
		if err != nil {
			return err
		}

		apiKey := viper.GetString("api-key")
		err = os.Setenv("FASTLY_API_KEY", apiKey)
		if err != nil {
			log.Fatal(err)
		}

		version, err := cmd.Flags().GetInt("version")
		if err != nil {
			return err
		}
		interactive, err := cmd.Flags().GetBool("interactive")
		if err != nil {
			return err
		}
		manageAll, err := cmd.Flags().GetBool("manage-all")
		if err != nil {
			return err
		}
		c := cli.Config{
			ID:          args[0],
			Version:     version,
			Directory:   workingDir,
			Interactive: interactive,
			ManageAll:   manageAll,
		}

		return importService(c)
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)

	// Persistent flags
	serviceCmd.PersistentFlags().IntP("version", "v", 0, "Version of the service to be imported")
	serviceCmd.PersistentFlags().BoolP("manage-all", "m", false, "Manage all associated resources")
}

func importService(c cli.Config) error {
	log.Printf("[INFO] Initializing Terraform")
	// Find/Install Terraform binary
	tf, err := terraform.Install(c.Directory)
	if err != nil {
		return err
	}

	// Create provider.tf
	// Create temp*.tf with empty service resource blocks
	log.Printf("[INFO] Creating provider.tf and temp*.tf")
	tempf, err := file.CreateInitTerraformFiles(c.Directory)
	defer os.Remove(tempf.Name())
	if err != nil {
		return err
	}

	// Run "terraform init"
	log.Printf(`[INFO] Running "terraform init"`)
	err = terraform.Init(tf)
	if err != nil {
		return err
	}

	// Run "terraform version"
	err = terraform.Version(tf)
	if err != nil {
		return err
	}

	// Create VCLServiceResourceProp struct
	serviceProp := prop.NewVCLServiceResource(c.ID, "service", c.Version)

	// log.Printf(`[INFO] Running "terraform import %s %s"`, serviceProp.GetRef(), serviceProp.GetIDforTFImport())
	log.Printf(`[INFO] Running "terraform import" on %s`, serviceProp.GetRef())
	err = terraform.Import(tf, serviceProp, tempf)
	if err != nil {
		return err
	}

	// Get the config represented in HCL from the "terraform show" output
	log.Print(`[INFO] Running "terraform show" to get the current Terraform state in HCL format`)
	rawHCL, err := terraform.Show(tf)

	// Parse HCL and obtain Terraform block props as a list of struct
	// to get the overall picture of the service configuration
	// log.Print("[INFO] Parsing the HCL to get an overall picture of the service configuration")
	log.Print("[INFO] Parsing the HCL")
	hcl, err := tfconf.Load(rawHCL)
	if err != nil {
		return err
	}

	props, err := hcl.ParseVCLServiceResource(serviceProp, c)
	if err != nil {
		return err
	}

	// Iterate over the list of props and run terraform import for WAF, ACL/dicitonary items, and dynamic snippets
	for _, p := range props {
		switch p := p.(type) {
		case *prop.WAFResource, *prop.ACLResource, *prop.DictionaryResource, *prop.DynamicSnippetResource:
			// Ask yes/no if in interactive mode
			if c.Interactive {
				yes := cli.YesNo(fmt.Sprintf("import %s? ", p.GetRef()))
				if !yes {
					continue
				}
			}

			log.Printf(`[INFO] Running "terraform import" on %s`, p.GetRef())
			terraform.Import(tf, p, tempf)
			if err != nil {
				return err
			}
		}
	}

	// temp*.tf no longer needed
	if err := tempf.Close(); err != nil {
		return err
	}
	if err := os.Remove(tempf.Name()); err != nil {
		return err
	}

	// Get the config represented in HCL from the "terraform show" output
	log.Print(`[INFO] Running "terraform show" to get the current Terraform state in HCL format`)
	rawHCL, err = terraform.Show(tf)

	// Make changes to the configuration
	// log.Print("[INFO] Parsing the HCL and making corrections removing read-only attrs and replacing embedded VCL/logformat with the file function")
	log.Print("[INFO] Parsing the HCL and making corrections")
	hcl, err = tfconf.Load(rawHCL)
	if err != nil {
		return err
	}

	sensitiveAttrs, err := hcl.RewriteResources(serviceProp, c)
	if err != nil {
		return err
	}

	log.Print("[INFO] Writing the configuration to main.tf")
	path := filepath.Join(c.Directory, "main.tf")
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	f.Write(hcl.Bytes())


	log.Print("[INFO] Writing .gitignore")
	if err := file.CreateGitIgnore(c.Directory); err != nil {
		return err
	}

	log.Print(`[INFO] Setting "activate" in terraform.tfstate`)
	curState, err := tfstate.Load(c.Directory)
	if err != nil {
		return err
	}

	newState, err := curState.SetActivateAttributes()
	if err != nil {
		return err
	}


	if c.ManageAll {
		log.Print(`[INFO] Settting "manage_*" in terraform.tfstate`)
		newState, err = newState.SetManageAttributes()
		if err != nil {
			return err
		}
	}

	for _, p := range props {
		switch p := p.(type) {
		case *prop.ACLResource, *prop.DictionaryResource, *prop.DynamicSnippetResource:
			log.Printf(`[INFO] Setting "index_key" in terraform.tfstate for %s`, p.GetRef())
			newState, err = newState.SetIndexKey(tfstate.SetIndexKeyParams{
				ResourceType: p.GetType(),
				ResourceName: p.GetNormalizedName(),
				Name:         p.GetName(),
			})
			if err != nil {
				return err
			}
		}
	}

	if len(sensitiveAttrs) > 0 {
		log.Print("[INFO] Writing variables.tf")
		variables := tfconf.BuildVariableDefinitions(sensitiveAttrs)
		if err := file.CreateVariablesTF(c.Directory, variables); err != nil {
			return err
		}

		log.Print("[INFO] Writing terraform.tfvars")
		tfvars := tfconf.BuildTFVars(sensitiveAttrs)
		if err := file.CreateTFVars(c.Directory, tfvars); err != nil {
			return err
		}

		log.Print(`[INFO] Setting "sensitive_attributes" in terraform.tfstate`)
		// Need to set once for each sensitive attribute
		blockTypes := map[string]struct{}{}
		for _, attr := range sensitiveAttrs {
			blockTypes[attr.BlockType] = struct{}{}
		}
		newState, err = newState.SetSensitiveAttributes(blockTypes)
		if err != nil {
			return err
		}
	}

	path = filepath.Join(c.Directory, "terraform.tfstate")
	f, err = os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0644)
	f.Write(newState.Bytes())
	f.Close()

	log.Print(`[INFO] Running "terraform refresh" to format the state file and check errors`)
	err = terraform.Refresh(tf)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, cli.BoldGreen("Completed!"))
	return nil
}
