package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
)

func Unzip(source, destination string) error {

	expandedDestination, err := homedir.Expand(destination)
	if err != nil {
		return err
	}

	r, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	if err := os.MkdirAll(expandedDestination, 0755); err != nil {
		return err
	}

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(expandedDestination, f.Name)
		if !strings.HasPrefix(path, filepath.Clean(expandedDestination)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: Illegal file path", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		fmt.Printf("Extracting file: %s to %s\n", f.Name, expandedDestination)
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
