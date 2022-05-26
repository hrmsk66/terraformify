package file

import (
	"errors"
	"os"
	"path/filepath"

	_ "embed"
)

const tfVersion = "1.1.9"

//go:embed static/provider.tf
var requiredProvider []byte

//go:embed static/.gitignore
var gitignore []byte

func CreateInitTerraformFiles(workingDir string) (*os.File, error) {
	// Create provider.tf
	if err := createProviderTF(workingDir); err != nil {
		return nil, err
	}

	// Create temp*.tf with empty service resource blocks
	tempf, err := os.CreateTemp(workingDir, "temp*.tf")
	if err != nil {
		return nil, err
	}

	return tempf, nil
}

func createProviderTF(workingDir string) error {
	return createFile(workingDir, "provider.tf", requiredProvider)
}

func CreateVariablesTF(workingDir string, content []byte) error {
	return createFile(workingDir, "variables.tf", content)
}

func CreateTFVars(workingDir string, content []byte) error {
	return createFile(workingDir, "terraform.tfvars", content)
}

func CreateGitIgnore(workingDir string) error {
	return createFile(workingDir, ".gitignore", gitignore)
}

func CreateContent(workingDir, name string, content []byte) error {
	return createFile(workingDir, name, content, "content")
}

func CreateVCL(workingDir, name string, content []byte) error {
	return createFile(workingDir, name, content, "vcl")
}

func CreateLogFormat(workingDir, name string, content []byte) error {
	return createFile(workingDir, name, content, "logformat")
}

func createFile(workingDir, name string, content []byte, ftypes ...string) error {
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
	return os.WriteFile(file, content, 0644)
}
