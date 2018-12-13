package libgin

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MakeZip creates a zip archive called dest from the files specified by source.
// Any directories listed in source are archived recursively.
func MakeZip(dest io.Writer, source ...string) error {
	// check sources
	for _, src := range source {
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("Cannot access '%s': %s", src, err.Error())
		}
	}

	zipwriter := zip.NewWriter(dest)
	defer zipwriter.Close()

	walker := func(filepath string, fi os.FileInfo, err error) error {

		// return on any error
		if err != nil {
			return err
		}

		// create a new dir/file header
		header, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when unzipping
		// header.Name = strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))
		header.Name = filepath

		if fi.Mode().IsDir() {
			return nil
		}

		// write the header
		w, err := zipwriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// Dereference symlinks
		if fi.Mode()&os.ModeSymlink != 0 {
			data, err := os.Readlink(filepath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(w, strings.NewReader(data)); err != nil {
				return err
			}
			return nil
		}

		// open files for zipping
		f, err := os.Open(filepath)
		defer f.Close()
		if err != nil {
			return err
		}

		// copy file data into zip writer
		if _, err := io.Copy(w, f); err != nil {
			return err
		}

		return nil
	}

	// walk path
	for _, src := range source {
		err := filepath.Walk(src, walker)
		if err != nil {
			return fmt.Errorf("Error adding %s to zip file: %s", src, err.Error())
		}
	}
	return nil
}
