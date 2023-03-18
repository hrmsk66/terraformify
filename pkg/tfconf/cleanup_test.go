package tfconf

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanupHCL(t *testing.T) {
	inputFile := filepath.Join("..", "..", "testdata", "cleanup_input.hcl")
	expectedOutputFile := filepath.Join("..", "..", "testdata", "cleanup_output.hcl")

	inputBytes, err := ioutil.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to read input file %s: %v", inputFile, err)
	}
	input := string(inputBytes)

	expectedOutputBytes, err := ioutil.ReadFile(expectedOutputFile)
	if err != nil {
		t.Fatalf("Failed to read expected output file %s: %v", expectedOutputFile, err)
	}
	expectedOutput := string(expectedOutputBytes)

	output := cleanupHCL(input)
	if strings.TrimSpace(output) != strings.TrimSpace(expectedOutput) {
		t.Errorf("cleanupHCL test failed.\nExpected:\n%v\n\nGot:\n%v", expectedOutput, output)
	}
}
