package terraform

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/hrmsk66/terraformify/pkg/prop"
)

func FindExec(workingDir string) (*tfexec.Terraform, error) {
	execPath, err := exec.LookPath("terraform")
	if err != nil {
		return nil, err
	}

	return tfexec.NewTerraform(workingDir, execPath)
}

func Init(tf *tfexec.Terraform) error {
	return tf.Init(context.Background(), tfexec.Upgrade(true))
}

func Version(tf *tfexec.Terraform) error {
	tfver, providerVers, err := tf.Version(context.Background(), true)
	if err != nil {
		return err
	}

	// Check if the version is compatible with 1.4.5 or earlier
	compatibleVersion, err := version.NewConstraint("<= 1.4.5")
	if err != nil {
		return fmt.Errorf("failed to parse version constraint: %s", err)
	}

	currentVersion, err := version.NewVersion(tfver.String())
	if err != nil {
		return fmt.Errorf("failed to parse current Terraform version: %s", err)
	}

	if !compatibleVersion.Check(currentVersion) {
		return fmt.Errorf("incompatible Terraform version: %s. Terraform version must be 1.4.5 or earlier", currentVersion)
	}

	log.Printf("[INFO] Terraform version: %s on %s_%s", currentVersion, runtime.GOOS, runtime.GOARCH)
	for k, v := range providerVers {
		log.Printf("[INFO] Provider version: %s %s", k, v.String())
	}
	return nil
}

func Import(tf *tfexec.Terraform, p prop.TFBlock, f io.Writer) error {
	// Add the empty resource block to the file
	_, err := fmt.Fprintf(f, "resource \"%s\" \"%s\" {}\n", p.GetType(), p.GetNormalizedName())
	if err != nil {
		return err
	}

	log.Printf(`[INFO] Running "terraform import" on %s`, p.GetRef())
	// Run "terraform import"
	if err := tf.Import(context.Background(), p.GetRef(), p.GetIDforTFImport()); err != nil {
		return err
	}

	return nil
}

// RecursiveImport attempts to import resources specified in the resource_link block of fastly_service_compute.
// As the resource_link lacks resource type information, this function iteratively tries to import using different
// resource types until it succeeds.
func RecursiveImport(tf *tfexec.Terraform, p prop.MutatableTfBlock, f io.Writer) error {
	err := Import(tf, p, f)

	if err != nil {
		// Importing non-existent data stores leads to varying outcomes based on their type:
		// - A non-existent fastly_configstore yields an error with "Cannot import non-existent remote object"
		// - A non-existent fastly_secretstore results in an error with "404 - Not Found"
		// - Surprisingly, a terraform import may succeed for a non-existent fastly_kvstore
		// To prevent erroneous state entries from non-existent resources, TFBlockProp.LinkedResource sequentially tries to import as:
		// "fastly_configstore" => "fastly_secretstore" => "fastly_kvstore".
		if strings.Contains(err.Error(), "Cannot import non-existent remote object") || strings.Contains(err.Error(), "404 - Not Found") {
			if mutateErr := p.MutateType(); mutateErr != nil {
				return mutateErr
			}
			log.Printf(`[INFO] - not found, retry with "%s"`, p.GetRef())
			return RecursiveImport(tf, p, f)
		}
		return err
	}
	return nil
}

func Show(tf *tfexec.Terraform) (string, error) {
	return tf.ShowPlanFileRaw(context.Background(), "terraform.tfstate")
}

func Refresh(tf *tfexec.Terraform) error {
	return tf.Refresh(context.Background())
}
