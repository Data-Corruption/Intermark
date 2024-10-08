package files

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
)

// Exists returns true if all given paths exist and false otherwise.
func Exists(paths ...string) (bool, error) {
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
	}
	return true, nil
}

// ListAllFiles returns a slice of all files in the directory and its subdirectories.
// If you don't want to include subdirectories, see ListFiles().
func ListAllFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // propagate the error up the call stack
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	} else {
		return files, nil
	}
}

// ListFiles returns a slice of all files in the directory.
// If you want to include subdirectories, see ListAllFiles().
func ListFiles(dir string, includeExt bool) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if !includeExt {
				name = name[:len(name)-len(filepath.Ext(name))]
			}
			files = append(files, name)
		}
	}
	return files, nil
}

// CopyFile copies the file at `src` to `dst`.
// If `dst` does not exist, it is created along with any missing parent directories.
// If `dst` exists, it is overwritten.
func CopyFile(src, dst string) error {
	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// CopyDir copies the directory at `src` to `dst`.
// If `dst` does not exist, it is created.
// If `dst` exists, it is deleted and replaced with `src`.
func CopyDir(src, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return err
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(dstPath, os.ModePerm)
		}
		return CopyFile(path, dstPath)
	})
}

func ReadFile(path string) (string, error) {
	if content, err := os.ReadFile(path); err != nil {
		return "", err
	} else {
		return string(content), nil
	}
}

func ReadFirstLine(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return scanner.Text(), nil
}

// CreateFile creates a file at the given path with the given content.
// If the file already exists, it is overwritten.
func CreateFile(path, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

// CleanDir removes then recreate the directory at the given path.
func CleanDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return os.MkdirAll(dir, os.ModePerm)
}

func EnsureDirs(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
