package xsd

import (
	"fmt"
	"os/exec"
)

// Validate validates the given XML file against the given XSD file using xmllint.
func Validate(xmlPath, xsdPath string) error {
	// Check if xmllint is available
	_, err := exec.LookPath("xmllint")
	if err != nil {
		return fmt.Errorf("xmllint not found: proper XSD validation requires xmllint installed on the system")
	}

	cmd := exec.Command("xmllint", "--noout", "--schema", xsdPath, xmlPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("validation failed:\n%s", string(output))
	}

	return nil
}
