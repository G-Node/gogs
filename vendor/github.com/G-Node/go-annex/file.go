package gannex

import (
	"os"
	"path/filepath"
	"strings"
	"io"
	"fmt"
	"regexp"
)

var (
	secPar    = regexp.MustCompile(`(\.\.)`)
	aFPattern = regexp.MustCompile(`[\\\/]annex[\\\/](.+)`)
)


type AFile struct {
	Filepath  string
	OFilename string
	Info      os.FileInfo
}

type AnnexFileNotFound struct {
	error
}

func NewAFile(annexpath, repopath, Ofilename string, APFileC []byte) (*AFile, error) {
	nAF := &AFile{OFilename: Ofilename}
	matches := aFPattern.FindStringSubmatch(string(APFileC))
	if matches != nil && len(matches) > 1 {
		filepath := strings.Replace(matches[1], "\\", "/", 0)
		filepath = fmt.Sprintf("%s/annex/%s", repopath, filepath)
		if !secPar.MatchString(filepath) {
			info, err := os.Stat(filepath)
			if err == nil {
				nAF.Filepath = filepath
				nAF.Info = info
				return nAF, nil
			}
		}
	}
	pathParts := strings.SplitAfter(string(APFileC), string(os.PathSeparator))
	filename := strings.TrimSpace(pathParts[len(pathParts)-1])
	// lets find the annex file
	filepath.Walk(filepath.Join(annexpath, repopath), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		} else if info.Name() == filename {
			nAF.Filepath = path
			nAF.Info = info
			return io.EOF
		}
		return nil
	})
	if nAF.Filepath != "" {
		return nAF, nil
	} else {
		return nil, AnnexFileNotFound{error: fmt.Errorf("Could not find File: %s anywhere below: %s", filename,
			filepath.Join(annexpath, repopath))}
	}

}

func (af *AFile) Open() (*os.File, error) {
	fp, err := os.Open(af.Filepath)
	if err != nil {
		return nil, err
	}
	return fp, nil

}
