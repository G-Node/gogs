// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/G-Node/gogs/pkg/setting"
)

type ArchiveType int

const (
	ZIP ArchiveType = iota + 1
	TARGZ
	GIN
)

func (c *Commit) CreateArchive(target string, archiveType ArchiveType, cloneL string) error {
	var format string
	switch archiveType {
	case ZIP:
		format = "zip"
	case TARGZ:
		format = "tar.gz"
	case GIN:
		to := filepath.Join(setting.Repository.Upload.TempPath, "archives", filepath.Base(strings.TrimSuffix(c.repo.Path, ".git")))
		defer os.RemoveAll(to)
		_, err := NewCommand("clone", c.repo.Path, to).RunTimeout(-1)
		if err != nil {
			return err
		}
		_, err = NewCommand("remote", "set-url", "origin", cloneL).RunInDirTimeout(-1, to)
		if err != nil {
			return err
		}
		fp, err := os.Create(target)
		defer fp.Close()
		if err != nil {
			return err
		}
		err = mkzip(to, fp)
		return err
	default:
		return fmt.Errorf("unknown format: %v", archiveType)
	}

	_, err := NewCommand("archive", "--prefix="+filepath.Base(strings.TrimSuffix(c.repo.Path, ".git"))+"/", "--format="+format, "-o", target, c.ID.String()).RunInDir(c.repo.Path)
	return err
}

// NOTE: TEMPORARY COPY FROM gin-doi

// mkzip walks the directory tree rooted at src and writes each file found to the given writers.
// The function accepts multiple writers to allow for multiple outputs (e.g., a file or md5 hash).
func mkzip(src string, writers ...io.Writer) error {
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
