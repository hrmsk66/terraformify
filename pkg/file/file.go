package file

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "embed"
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

func WriteMainTF(workingDir string, content []byte) error {
	return writeFile(workingDir, "main.tf", content)
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

func WriteContent(workingDir, name string, content []byte) error {
	return writeFile(workingDir, name, content, "content")
}

func WriteVCL(workingDir, name string, content []byte) error {
	return writeFile(workingDir, name, content, "vcl")
}

func WriteLogFormat(workingDir, name string, content []byte) error {
	return writeFile(workingDir, name, content, "logformat")
}

func writeFile(workingDir, name string, content []byte, ftypes ...string) error {
	for _, ftype := range ftypes {
		dir := filepath.Join(workingDir, ftype)
		if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(dir, 0755)
			if err != nil {
				return err
			}
		}

		workingDir = filepath.Join(workingDir, ftype)
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
	if name == "main.tf" || name == "variables.tf" || name == "terraform.tfvars" {
		log.Printf("[INFO] file: %s exists. appending content", file)
		return write(file, content, os.O_WRONLY|os.O_APPEND)
	}
	// Overwrite
	if name == "terraform.tfstate" {
		log.Print("[INFO] file: writing terraform.tfstate")
		return write(file, content, os.O_WRONLY|os.O_TRUNC)
	}

	return fmt.Errorf("aborted creating %s, because it already exists", file)
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
