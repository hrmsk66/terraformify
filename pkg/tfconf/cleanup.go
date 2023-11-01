package tfconf

import (
	"bufio"
	"bytes"
	"strings"
)

func cleanupHCL(rawHCL string) string {
	var buf bytes.Buffer
	scanner := bufio.NewScanner(strings.NewReader(rawHCL))

	var blocks []string
	var eot string
	var skip bool

	// Helper function to handle multiline strings
	handleMultilineStrings := func(t string) {
		switch {
		case strings.HasSuffix(t, "EOT"):
			skip = true
			eot = "EOT"
		case strings.HasSuffix(t, "jsonencode("):
			skip = true
			eot = ")"
		}
	}

	// Helper function to truncate attribute values
	truncateValue := func(t string) string {
		before, _, _ := strings.Cut(t, "=")
		return before + `= ""`
	}

	// Helper function to check if a given block is a supported Terraform resource block
	isSupportedResourceBlock := func(block string) bool {
		supportedBlocks := []string{
			"fastly_service_acl_entries",
			"fastly_service_compute",
			"fastly_service_dictionary_items",
			"fastly_service_dynamic_snippet_content",
			"fastly_service_vcl",
			"fastly_service_waf_configuration",
		}

		for _, supportedBlock := range supportedBlocks {
			if strings.HasPrefix(block, supportedBlock) {
				return true
			}
		}
		return false
	}

	for scanner.Scan() {
		text := scanner.Text()
		trimedText := strings.TrimSpace(scanner.Text())

		// Skip lines until the end of the multiline string is found
		if skip {
			if trimedText == eot {
				skip = false
			}
			continue
		}

		// Check for the opening bracket of a block
		if strings.HasSuffix(trimedText, "{") {
			switch len(blocks) {
			case 0:
				// If we're not in a block, check if it's a supported resource block
				if strings.HasPrefix(trimedText, "resource") {
					b := strings.Fields(trimedText)[1]
					b = strings.Trim(b, "\"")
					if isSupportedResourceBlock(b) {
						blocks = append(blocks, b)
					}
				}
			case 1:
				// If we're inside a resource block, check if the block needs a special handling
				switch {
				case strings.HasPrefix(trimedText, "backend"):
					blocks = append(blocks, "backend")
				case strings.HasPrefix(trimedText, "response_object"):
					blocks = append(blocks, "response_object")
				case strings.HasPrefix(trimedText, "snippet"):
					blocks = append(blocks, "snippet")
				case strings.HasPrefix(trimedText, "vcl"):
					blocks = append(blocks, "vcl")
				case strings.HasPrefix(trimedText, "logging_"):
					blocks = append(blocks, "logging")
				default:
					blocks = append(blocks, "other")
				}
			}
		}

		// Continue if we're not inside a supported Terraform resource block
		if len(blocks) == 0 {
			continue
		}

		// If we find a closing bracket, remove the current block from the list
		if trimedText == "}" {
			blocks = blocks[:len(blocks)-1]
		}

		// Special handling for nested blocks
		if len(blocks) > 0 {
			if blocks[len(blocks)-1] == "fastly_service_dynamic_snippet_content" {
				switch {
				case strings.HasPrefix(trimedText, "content "):
					handleMultilineStrings(trimedText)
					text = truncateValue(text)
				}
			}

			if blocks[len(blocks)-1] == "backend" {
				switch {
				case strings.HasSuffix(trimedText, "(sensitive value)"):
					text = truncateValue(text)
				}
			}

			if blocks[len(blocks)-1] == "response_object" {
				switch {
				case strings.HasPrefix(trimedText, "content "):
					handleMultilineStrings(trimedText)
					text = truncateValue(text)
				}
			}

			if blocks[len(blocks)-1] == "snippet" {
				switch {
				case strings.HasPrefix(trimedText, "content "):
					handleMultilineStrings(trimedText)
					text = truncateValue(text)
				}
			}

			if blocks[len(blocks)-1] == "vcl" {
				switch {
				case strings.HasPrefix(trimedText, "content "):
					handleMultilineStrings(trimedText)
					text = truncateValue(text)
				}
			}

			if blocks[len(blocks)-1] == "logging" {
				switch {
				case strings.HasPrefix(trimedText, "format "):
					handleMultilineStrings(trimedText)
					text = truncateValue(text)
				case strings.HasSuffix(trimedText, "(sensitive value)"):
					text = truncateValue(text)
				}
			}
		}

		buf.WriteString(text + "\n")
	}

	return buf.String()
}
