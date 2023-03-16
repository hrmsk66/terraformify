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

	// helper fn; set `skip` and `eot` if we are handling multiline strings
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

	// helper fn; truncate attribute value
	truncateValue := func(t string) string {
		before, _, _ := strings.Cut(t, "=")
		return before + `= ""`
	}

	for scanner.Scan() {
		text := scanner.Text()
		trimedText := strings.TrimSpace(scanner.Text())

		// Skip until eot
		if skip {
			if trimedText == eot {
				skip = false
			}
			continue
		}

		// Set the block name if the block has attributes that needs to be truncated
		if strings.HasSuffix(trimedText, "{") {
			switch {
			case strings.HasPrefix(trimedText, "resource"):
				blocks = append(blocks, "resource")
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

		// Skip as we are not in a TF resource block
		if len(blocks) == 0 {
			continue
		}

		// Closing bracket; clear the block name
		if trimedText == "}" {
			blocks = blocks[:len(blocks) - 1]
		}

		if len(blocks) > 0 {
			if blocks[len(blocks) - 1] == "backend" {
				switch {
				case strings.HasSuffix(trimedText, "(sensitive value)"):
					text = truncateValue(text)
				}
			}

			if blocks[len(blocks) - 1] == "response_object" {
				switch {
				case strings.HasPrefix(trimedText, "content "):
					handleMultilineStrings(trimedText)
					text = truncateValue(text)
				}
			}

			if blocks[len(blocks) - 1] == "snippet" {
				switch {
				case strings.HasPrefix(trimedText, "content "):
					handleMultilineStrings(trimedText)
					text = truncateValue(text)
				}
			}

			if blocks[len(blocks) - 1] == "vcl" {
				switch {
				case strings.HasPrefix(trimedText, "content "):
					handleMultilineStrings(trimedText)
					text = truncateValue(text)
				}
			}

			if blocks[len(blocks) - 1] == "logging" {
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
