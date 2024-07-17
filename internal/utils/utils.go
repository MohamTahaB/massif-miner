package utils

import (
	"fmt"
	"regexp"
)

// Extracts the value from a line of the form "label=value"
// Returns the value or (xor) an error if not found
func ExtractValueOf(label string, line string, valueIsANumber bool) (string, error) {

	// Define whether the value is a number or a string
	valueRegex := "."
	if valueIsANumber {
		valueRegex = `\d`
	}
	pattern := fmt.Sprintf(`%s=(%s+)$`, label, valueRegex)
	regex := regexp.MustCompile(pattern)

	stringSubmatches := regex.FindStringSubmatch(line)

	if len(stringSubmatches) < 2 {
		return "", fmt.Errorf("unexpected regex issue")
	}

	return stringSubmatches[1], nil
}
