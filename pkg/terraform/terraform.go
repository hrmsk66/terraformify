package terraform

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"runtime"

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

	// Run "terraform import"
	if err := tf.Import(context.Background(), p.GetRef(), p.GetIDforTFImport()); err != nil {
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
