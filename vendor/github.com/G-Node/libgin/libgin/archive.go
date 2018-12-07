package libgin

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MkZip walks the directory tree rooted at src and writes each file found to the given writers.
// The function accepts multiple writers to allow for multiple outputs (e.g., a file or md5 hash).
func MkZip(src string, writers ...io.Writer) error {
	// ensure the src actually exists before trying to zip it
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("Unable to zip files: %s", err.Error())
	}

	mw := io.MultiWriter(writers...)

	tw := zip.NewWriter(mw)
	defer tw.Close()

	// walk path
	return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

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
		header.Name = strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))

		// write the header
		w, err := tw.CreateHeader(header)
		if err != nil {
			return err
		}

		// return on directories since there will be no content to zip
		if fi.Mode().IsDir() {
			return nil
		}
		mode := fi.Mode()
		fmt.Print(mode)
		if fi.Mode()&os.ModeSymlink != 0 {
			data, err := os.Readlink(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(w, strings.NewReader(data)); err != nil {
				return err
			}
			return nil
		}

		// open files for zipping
		f, err := os.Open(file)
		defer f.Close()
		if err != nil {
			return err
		}

		// copy file data into zip writer
		if _, err := io.Copy(w, f); err != nil {
			return err
		}

		return nil
	})
}
