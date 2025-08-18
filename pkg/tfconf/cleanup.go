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
	var skipSensitiveBlock bool
	var sensitiveBlockType string
	var sensitiveBlockDepth int
	var skippedBlocks []string

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

	// Helper function to detect entirely sensitive blocks
	isSensitiveBlockComment := func(t string) bool {
		return strings.Contains(t, "# At least one attribute in this block is (or was) sensitive") ||
			strings.Contains(t, "# so its contents will not be displayed")
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
			"fastly_configstore",
			"fastly_configstore_entries",
			"fastly_secretstore",
			"fastly_kvstore",
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

		// Handle skipping entirely sensitive blocks
		if skipSensitiveBlock {
			// Count braces to track when we're out of the sensitive block
			if strings.HasSuffix(trimedText, "{") {
				sensitiveBlockDepth++
			} else if trimedText == "}" {
				sensitiveBlockDepth--
				if sensitiveBlockDepth == 0 {
					// We're out of the sensitive block - also pop from blocks array
					skipSensitiveBlock = false
					skippedBlocks = append(skippedBlocks, sensitiveBlockType)
					sensitiveBlockType = ""
					// Pop the block we just finished from the blocks stack
					if len(blocks) > 0 {
						blocks = blocks[:len(blocks)-1]
					}
					continue // Skip the closing brace too
				}
			}
			continue // Skip all content within sensitive blocks
		}

		// Detect the start of an entirely sensitive block
		if len(blocks) >= 2 && isSensitiveBlockComment(trimedText) {
			// We found a sensitive block comment - this means the block we just entered is sensitive
			// Start skipping from the next line (we don't write the block header)
			skipSensitiveBlock = true
			sensitiveBlockType = blocks[len(blocks)-1]
			sensitiveBlockDepth = 1
			
			// Remove the block header from the buffer by trimming the last few lines
			bufStr := buf.String()
			lines := strings.Split(bufStr, "\n")
			
			// Remove the block header line (should be the last non-empty line)
			for i := len(lines) - 1; i >= 0; i-- {
				line := strings.TrimSpace(lines[i])
				if line != "" && (strings.HasSuffix(line, "{") && 
					(strings.Contains(line, sensitiveBlockType) || 
					 (sensitiveBlockType == "logging" && strings.Contains(line, "logging_")))) {
					// Found the block header to remove
					lines = lines[:i]
					break
				}
			}
			
			// Rebuild the buffer
			buf.Reset()
			if len(lines) > 0 {
				buf.WriteString(strings.Join(lines, "\n"))
				buf.WriteString("\n")
			}
			
			continue
		}

		// Check for empty line and preserve it
		if trimedText == "" {
			buf.WriteString(text + "\n")
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
				var blockType string
				switch {
				case strings.HasPrefix(trimedText, "backend"):
					blockType = "backend"
				case strings.HasPrefix(trimedText, "rate_limiter"):
					blockType = "rate_limiter"
				case strings.HasPrefix(trimedText, "response_object"):
					blockType = "response_object"
				case strings.HasPrefix(trimedText, "snippet"):
					blockType = "snippet"
				case strings.HasPrefix(trimedText, "vcl"):
					blockType = "vcl"
				case strings.HasPrefix(trimedText, "logging_"):
					blockType = "logging"
				default:
					blockType = "other"
				}
				blocks = append(blocks, blockType)
			case 2:
				// If we're inside "rate_limiter" block, check if it's a "response" block
				if blocks[len(blocks)-1] == "rate_limiter" && strings.HasPrefix(trimedText, "response ") {
					blocks = append(blocks, "response")
				}
				// If we're inside "product_enablement" block, check if it's a "ngwaf" block
				if blocks[len(blocks)-1] == "other" && strings.HasPrefix(trimedText, "ngwaf ") {
					blocks = append(blocks, "ngwaf")
				}
			}
		}

		// Continue if we're not inside a supported Terraform resource block
		if len(blocks) == 0 {
			continue
		}

		// If we find a closing bracket, remove the current block from the list
		// (but only if we're not currently skipping a sensitive block)
		if trimedText == "}" && !skipSensitiveBlock {
			if len(blocks) > 0 {
				blocks = blocks[:len(blocks)-1]
			}
		}

		// Special handling for nested blocks
		switch len(blocks) {
		case 1:
			if blocks[len(blocks)-1] == "fastly_service_dynamic_snippet_content" {
				switch {
				case strings.HasPrefix(trimedText, "content "):
					handleMultilineStrings(trimedText)
					text = truncateValue(text)
				}
			}
		case 2:
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
		case 3:
			if blocks[len(blocks)-2] == "rate_limiter" && blocks[len(blocks)-1] == "response" {
				switch {
				case strings.HasPrefix(trimedText, "content "):
					handleMultilineStrings(trimedText)
					text = truncateValue(text)
				}
			}
		}

		buf.WriteString(text + "\n")
	}

	// Add documentation comments for all skipped sensitive blocks
	if len(skippedBlocks) > 0 {
		buf.WriteString("\n")
		buf.WriteString("# =============================================================================\n")
		buf.WriteString("# SENSITIVE BLOCKS REMOVED\n")
		buf.WriteString("# =============================================================================\n")
		buf.WriteString("# The following blocks were removed due to sensitive values and need manual configuration:\n")
		for _, blockType := range skippedBlocks {
			buf.WriteString("# - " + blockType + " block\n")
		}
		buf.WriteString("#\n")
		buf.WriteString("# This is due to enhanced security in Fastly Terraform Provider v8.0.0+\n")
		buf.WriteString("# where entire blocks are marked as sensitive when they contain sensitive attributes.\n")
		buf.WriteString("#\n")
		buf.WriteString("# To complete your configuration:\n")
		buf.WriteString("# 1. Refer to the Fastly Terraform provider documentation\n")
		buf.WriteString("# 2. Add the required blocks manually with appropriate attributes\n")
		buf.WriteString("# 3. Use Terraform variables for sensitive values\n")
		buf.WriteString("#\n")
		buf.WriteString("# Documentation: https://registry.terraform.io/providers/fastly/fastly/latest/docs\n")
		buf.WriteString("# =============================================================================\n")
	}

	return buf.String()
}
