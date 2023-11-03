package file

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/hrmsk66/terraformify/pkg/cli"
)

//go:embed static/provider.tf
var requiredProvider []byte

//go:embed static/.gitignore
var gitignore []byte

func CreateInitTerraformFiles(workingDir string) (*os.File, error) {
	// Create provider.tf
	if err := writeProviderTF(workingDir); err != nil {
		return nil, err
	}

	// Create temp*.tf with empty service resource blocks
	tempf, err := os.CreateTemp(workingDir, "temp*.tf")
	if err != nil {
		return nil, err
	}

	return tempf, nil
}

func WriteTF(workingDir, resourceName string, content []byte) error {
	filename := fmt.Sprintf("%s.tf", resourceName)
	return writeFile(workingDir, filename, content)
}

func WriteTFState(workingDir string, content []byte) error {
	return writeFile(workingDir, "terraform.tfstate", content)
}

func writeProviderTF(workingDir string) error {
	lockFile := filepath.Join(workingDir, ".terraform.lock.hcl")
	_, err := os.Stat(lockFile)
	if errors.Is(err, os.ErrNotExist) {
		return writeFile(workingDir, "provider.tf", requiredProvider)
	}
	if err != nil {
		return err
	}

	log.Printf("[INFO] file: %s exists. skip creating provider.tf", lockFile)
	return nil
}

func WriteVariablesTF(workingDir string, content []byte) error {
	return writeFile(workingDir, "variables.tf", content)
}

func WriteTFVars(workingDir string, content []byte) error {
	return writeFile(workingDir, "terraform.tfvars", content)
}

func WriteGitIgnore(workingDir string) error {
	return writeFile(workingDir, ".gitignore", gitignore)
}

func WriteContent(workingDir, resourceName, fileName string, content []byte) error {
	return writeFile(workingDir, fileName, content, "content", resourceName)
}

func WriteVCL(workingDir, resourceName, fileName string, content []byte) error {
	return writeFile(workingDir, fileName, content, "vcl", resourceName)
}

func WriteLogFormat(workingDir, resourceName, fileName string, content []byte) error {
	return writeFile(workingDir, fileName, content, "logformat", resourceName)
}

func writeFile(workingDir, name string, content []byte, dirs ...string) error {
	for _, dir := range dirs {
		d := filepath.Join(workingDir, dir)
		if _, err := os.Stat(d); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(d, 0755)
			if err != nil {
				return err
			}
		}

		workingDir = filepath.Join(workingDir, dir)
	}

	file := filepath.Join(workingDir, name)
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("[INFO] file: creating %s", file)
		return write(file, content, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	}
	if err != nil {
		return err
	}
	// Skip
	if name == "provider.tf" || name == ".gitignore" {
		log.Printf("[INFO] file: %s exists. skip creating it", file)
		return nil
	}
	// Append
	if name == "variables.tf" || name == "terraform.tfvars" {
		log.Printf("[INFO] file: %s exists. appending content", file)
		return write(file, content, os.O_WRONLY|os.O_APPEND)
	}
	// Overwrite
	if name == "terraform.tfstate" {
		log.Print("[INFO] file: writing terraform.tfstate")
		return write(file, content, os.O_WRONLY|os.O_TRUNC)
	}

	return fmt.Errorf("aborted creating %s as it already exists", file)
}

func write(file string, content []byte, flag int) error {
	f, err := os.OpenFile(file, flag, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

// CheckPackage verifies that the file at the given path exists and is a .tar.gz file.
func CheckPackage(packagePath string) error {
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return err
	}

	if filepath.Ext(strings.TrimSuffix(packagePath, filepath.Ext(packagePath))) != ".tar" || filepath.Ext(packagePath) != ".gz" {
		return errors.New("the specified file is not in tar.gz format")
	}

	return nil
}

func CheckFile(workingDir, resourceName string) error {
	fileName := fmt.Sprintf("%s.tf", resourceName)
	file := filepath.Join(workingDir, fileName)

	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	return fmt.Errorf("aborted creating a TF config file named %#v as it already exists. try another TF resource name using the -n option (the file is named after the TF resource name and it defaults to service.tf)", file)
}

func CheckDir(workingDir string, autoYes bool) (err error) {
	info, err := os.Stat(workingDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", workingDir)
	}

	d, err := os.Open(workingDir)
	if err != nil {
		return err
	}

	defer func() {
		if err1 := d.Close(); err1 != nil {
			err = err1
		}
	}()

	_, err = d.Readdir(1)
	if err == io.EOF {
		return nil
	}

	msg := `WARNING: Working Directory Not Empty
   The working directory is not empty.
   If the import fails, the files in the directory may be left in an inconsistent state.
   Please ensure that you back up the directory before proceeding.
   Do you want to continue?`
	if autoYes || cli.YesNo(msg) {
		return nil
	}
	return errors.New("working directory is not empty")
}
