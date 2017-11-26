package gindex

import (
	"bufio"

	"net/http"
	"strings"

	"github.com/G-Node/gogs/pkg/tool"
	"github.com/Sirupsen/logrus"
)

const (
	UKKNOWN = iota
	ANNEX
	ODML_XML
	TEXT
)

func DetermineFileType(peekData []byte) (int64, error) {
	if tool.IsAnnexedFile(peekData){
		logrus.Debugf("Found an annex file")
		return ANNEX,nil
	}
	typeStr := http.DetectContentType(peekData)
	if strings.Contains(typeStr, "text") {
		if strings.Contains(string(peekData), "ODML") {
			return ODML_XML, nil
		}
		logrus.Debugf("Found a text file")
		return TEXT, nil
	}
	return UKKNOWN, nil

}
func BlobFileType(blob *IndexBlob) (int64, *bufio.Reader, error) {
	blobBuffer := bufio.NewReader(blob.Blob)
	if blob.Size() > 1024 {
		peekData, err := blobBuffer.Peek(1024)
		if err != nil {
			return UKKNOWN,nil, err
		}
		fType, err := DetermineFileType(peekData)
		return fType, blobBuffer, err
	} else {
		peekData, err := blobBuffer.Peek(int(blob.Size())) // conversion should be fine(<1024)
		if err != nil {
			return UKKNOWN, nil, err
		}
		fType, err := DetermineFileType(peekData)
		return fType, blobBuffer, err
	}

}
