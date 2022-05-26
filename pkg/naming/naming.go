package naming

import (
	"regexp"
	"strings"
)

func Normalize(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "\n", "_")
	name = strings.ReplaceAll(name, "\t", "_")
	return strings.ReplaceAll(name, " ", "_")
}

func IsValid(name string) bool {
	// Validate if the string can be used as a Terraform resource name
	// - No check is necessary for "fastly_service_waf_configuration" because the name is fixed to "waf"
	// - No check is necessary for the following resources because invalid names are not accepted at Fastly
	//	- "fastly_service_acl_entries"
	//	- "fastly_service_dictionary_items"
	//	- "fastly_service_dynamic_snippet_content"

	// A TF resource names begin with a letter or underscore and may contain only letters, digits, underscores, and dashes
	// Spaces and dots are allowed here since they are replaced with underscores in TFBlockProp.GetNormalizedName()
	return regexp.MustCompile(`^[A-Za-z_][0-9A-Za-z_.\-\\s]*$`).MatchString(name)
}