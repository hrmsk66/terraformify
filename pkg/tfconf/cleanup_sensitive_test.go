package tfconf

import (
	"strings"
	"testing"
)

func TestCleanupHCLWithSensitiveBlocks(t *testing.T) {
	input := `# fastly_service_vcl.service:
resource "fastly_service_vcl" "service" {
    active_version     = 30
    cloned_version     = 30
    comment            = "Managed by Terraform"
    id                 = "6Ns1vcUDdC1eSjEJiHFXVE"
    name               = "demofastly"

    backend {
      # At least one attribute in this block is (or was) sensitive,
      # so its contents will not be displayed.
    }
    backend {
      # At least one attribute in this block is (or was) sensitive,
      # so its contents will not be displayed.
    }

    domain {
        comment = null
        name    = "testz12.com"
    }

    logging_ftp {
      # At least one attribute in this block is (or was) sensitive,
      # so its contents will not be displayed.
    }

    logging_gcs {
      # At least one attribute in this block is (or was) sensitive,
      # so its contents will not be displayed.
    }

    product_enablement {
        bot_management = false
        image_optimizer = true
    }
}
`

	result := cleanupHCL(input)

	// Check that sensitive blocks are removed
	if strings.Contains(result, "backend {") {
		t.Error("Expected backend blocks to be removed, but they are still present")
	}

	if strings.Contains(result, "logging_ftp {") {
		t.Error("Expected logging_ftp blocks to be removed, but they are still present")
	}

	if strings.Contains(result, "logging_gcs {") {
		t.Error("Expected logging_gcs blocks to be removed, but they are still present")
	}

	// Check that non-sensitive blocks are preserved
	if !strings.Contains(result, "domain {") {
		t.Error("Expected domain block to be preserved")
	}

	if !strings.Contains(result, "product_enablement {") {
		t.Error("Expected product_enablement block to be preserved")
	}

	// Check that documentation is added
	if !strings.Contains(result, "SENSITIVE BLOCKS REMOVED") {
		t.Error("Expected documentation about removed sensitive blocks")
	}

	if !strings.Contains(result, "backend block") {
		t.Error("Expected documentation to mention removed backend blocks")
	}

	if !strings.Contains(result, "logging block") {
		t.Error("Expected documentation to mention removed logging blocks")
	}

	// Check that the sensitive comment detection pattern is not in output
	if strings.Contains(result, "At least one attribute in this block is (or was) sensitive") {
		t.Error("Expected sensitive comments to be removed from output")
	}

	t.Logf("Cleaned HCL result:\n%s", result)
}

func TestCleanupHCLWithNoSensitiveBlocks(t *testing.T) {
	input := `# fastly_service_vcl.service:
resource "fastly_service_vcl" "service" {
    comment = "Test service"
    name    = "test"

    domain {
        name = "example.com"
    }

    backend {
        name    = "example"
        address = "example.com"
    }
}
`

	result := cleanupHCL(input)

	// Check that all blocks are preserved when not sensitive
	if !strings.Contains(result, "backend {") {
		t.Error("Expected backend block to be preserved when not sensitive")
	}

	if !strings.Contains(result, "domain {") {
		t.Error("Expected domain block to be preserved")
	}

	// Check that no documentation is added when no blocks are removed
	if strings.Contains(result, "SENSITIVE BLOCKS REMOVED") {
		t.Error("Expected no documentation when no blocks are removed")
	}

	t.Logf("Cleaned HCL result:\n%s", result)
}