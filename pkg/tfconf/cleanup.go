package tfconf

import (
	"bufio"
	"bytes"
	"strings"
)

func cleanupHCL(rawHCL string) string {
	var buf bytes.Buffer
	scanner := bufio.NewScanner(strings.NewReader(rawHCL))

	var block, eot string
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

		// Closing bracket; clear the block name
		if block != "" && trimedText == "}" {
			block = ""
		}

		// Set the block name if the block has attributes that needs to be truncated
		if block == "" && strings.HasSuffix(trimedText, "{") {
			switch {
			case strings.HasPrefix(trimedText, "backend"):
				block = "backend"
			case strings.HasPrefix(trimedText, "response_object"):
				block = "response_object"
			case strings.HasPrefix(trimedText, "snippet"):
				block = "snippet"
			case strings.HasPrefix(trimedText, "vcl"):
				block = "vcl"
			case strings.HasPrefix(trimedText, "logging_"):
				block = "logging"
			}
		}

		if block == "backend" {
			switch {
			case strings.HasSuffix(trimedText, "(sensitive value)"):
				text = truncateValue(text)
			}
		}

		if block == "response_object" {
			switch {
			case strings.HasPrefix(trimedText, "content "):
				handleMultilineStrings(trimedText)
				text = truncateValue(text)
			}
		}

		if block == "snippet" {
			switch {
			case strings.HasPrefix(trimedText, "content "):
				handleMultilineStrings(trimedText)
				text = truncateValue(text)
			}
		}

		if block == "vcl" {
			switch {
			case strings.HasPrefix(trimedText, "content "):
				handleMultilineStrings(trimedText)
				text = truncateValue(text)
			}
		}

		if block == "logging" {
			switch {
			case strings.HasPrefix(trimedText, "format "):
				handleMultilineStrings(trimedText)
				text = truncateValue(text)
			case strings.HasSuffix(trimedText, "(sensitive value)"):
				text = truncateValue(text)
			}
		}

		buf.WriteString(text + "\n")
	}

	return buf.String()
}
