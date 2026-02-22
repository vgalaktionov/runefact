package build

import (
	"fmt"
	"os"
	"unicode/utf8"
)

// checkUTF8 validates that a file contains valid UTF-8 text.
func checkUTF8(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !utf8.Valid(data) {
		return fmt.Errorf("%s: file is not valid UTF-8, please re-encode", path)
	}
	return nil
}
