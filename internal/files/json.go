package files

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
)

// LoadJSON loads the JSON from the given file path into the given object.
// If the file does not exist, (false, nil) is returned.
func LoadJSON(path string, obj interface{}) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error checking file: %w", err)
	}
	file, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("error reading file: %w", err)
	}
	if err := json.Unmarshal(file, obj); err != nil {
		return false, fmt.Errorf("error parsing file: %w", err)
	}
	return true, nil
}

// SaveJSON saves the object to the given file path with the given permissions.
func SaveJSON(path string, obj interface{}, perms fs.FileMode) error {
	data, err := json.MarshalIndent(obj, "", "  ") // with indentation
	if err != nil {
		return fmt.Errorf("error encoding obj: %w", err)
	}
	if err = os.WriteFile(path, data, perms); err != nil {
		return fmt.Errorf("error writing obj file: %w", err)
	}
	return nil
}
