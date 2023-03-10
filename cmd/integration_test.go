package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hrmsk66/terraformify/pkg/cli"
	"github.com/stretchr/testify/require"
)

const domain = "test.terraformify.me"
const prepDir = "tmfy-test-prep"
const testDir = "tfmy-test"
const packageFile = "package.tar.gz"

func TestMain(m *testing.M) {
	setup()
	defer cleanup()

	m.Run()
}

func setup() {
	prepDirPath := filepath.Join(os.TempDir(), prepDir)
	_ = os.Mkdir(prepDirPath, 0700)
	provider, _ := os.ReadFile("../testdata/provider.tf")
	providerPath := filepath.Join(prepDirPath, "provider.tf")
	os.WriteFile(providerPath, provider, 0644)
}

func cleanup() {
	prepDirPath := filepath.Join(os.TempDir(), prepDir)
	os.RemoveAll(prepDirPath)
}

// prep deploys a service that terraformify will import in the test
func prep(t *testing.T, files ...string) (*terraform.Options, error) {
	t.Logf("preparing for %s", t.Name())
	defer t.Logf("preparation completed for %s", t.Name())

	main, err := os.ReadFile("../testdata/" + files[0])
	if err != nil {
		return nil, err
	}
	prepDirPath := filepath.Join(os.TempDir(), prepDir)

	// Remove terraform.tfstate from previous test
	// os.Remove returns an error in the first test item because terraform.tfstate doesn't exist, which is not a problem.
	statePath := filepath.Join(prepDirPath, "terraform.tfstate")
	if err := os.Remove(statePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	// Write/rewrite config in main.tf
	mainPath := filepath.Join(prepDirPath, "main.tf")
	if err := os.WriteFile(mainPath, main, 0644); err != nil {
		return nil, err
	}

	// Copy C@E package file to the test directory
	if len(files) == 2 {
		pkg, err := os.ReadFile("../testdata/" + files[1])
		if err != nil {
			return nil, err
		}

		pkgPath := filepath.Join(prepDirPath, packageFile)
		if err := os.WriteFile(pkgPath, pkg, 0644); err != nil {
			return nil, err
		}
	}

	opt := terraform.WithDefaultRetryableErrors(
		t,
		&terraform.Options{
			TerraformDir: prepDirPath,
			Vars: map[string]interface{}{
				"domain": domain,
			},
		},
	)

	if _, err := terraform.InitAndApplyE(t, opt); err != nil {
		return nil, err
	}
	return opt, nil
}

// The test cases are not likely to be completed in 10 mins. Run with `-timeout 30m`
func TestImportService(t *testing.T) {
	testCases := []struct {
		resourceType     string
		name             string
		expResourceCount int
	}{
		{"vcl", "service_custom_vcl.tf", 1},
		{"vcl", "service_acl.tf", 3},
		{"vcl", "service_dictionary.tf", 3},
		{"vcl", "service_dynamic_snippet.tf", 3},
		{"vcl", "service_waf.tf", 2},
		{"compute", "service_compute.tf", 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var prepOpt *terraform.Options
			var err error

			if tc.resourceType == "vcl" {
				prepOpt, err = prep(t, tc.name)
				if err != nil {
					t.Errorf("Failed to set up a test service: %s", err)
				}
			}

			if tc.resourceType == "compute" {
				prepOpt, err = prep(t, tc.name, packageFile)
				if err != nil {
					t.Errorf("Failed to set up a test service: %s", err)
				}
			}

			// Create a working directory
			testDirPath, err := os.MkdirTemp("", testDir)
			if err != nil {
				t.Errorf("Failed to create a working directory: %s", err)
			}
			defer os.RemoveAll(testDirPath)

			// Read the service ID from the result of prep()
			serviceID := terraform.Output(t, prepOpt, "id")
			c := cli.Config{
				ID:           serviceID,
				Package:      packageFile,
				Directory:    testDirPath,
				ForceDestroy: true,
			}

			if tc.resourceType == "vcl" {
				// Run terraformify
				if err = importVCL(c); err != nil {
					t.Errorf("Failed to import the service: %s", err)
				}
			}
			if tc.resourceType == "compute" {
				pkg, err := os.ReadFile("../testdata/" + packageFile)
				if err != nil {
					t.Errorf("Failed to read %s", packageFile)
				}

				pkgPath := filepath.Join(testDirPath, packageFile)
				if err := os.WriteFile(pkgPath, pkg, 0644); err != nil {
					t.Errorf("Failed to write %s", packageFile)
				}

				// Run terraformify
				if err = importCompute(c); err != nil {
					t.Errorf("Failed to import the service: %s", err)
				}
			}

			// Run "terraform apply". add/change/destroy counts should all be 0
			testOpt := terraform.WithDefaultRetryableErrors(
				t,
				&terraform.Options{
					TerraformDir: testDirPath,
				},
			)
			applyString := terraform.Apply(t, testOpt)
			applyCounts := terraform.GetResourceCount(t, applyString)
			require.Equal(t, 0, applyCounts.Add)
			require.Equal(t, 0, applyCounts.Change)
			require.Equal(t, 0, applyCounts.Destroy)

			// Run "terraform destroy". destroy counts should match expResourceCount
			destroyString := terraform.Destroy(t, testOpt)
			destroyCounts := terraform.GetResourceCount(t, destroyString)
			require.Equal(t, tc.expResourceCount, destroyCounts.Destroy)
		})
	}
}
