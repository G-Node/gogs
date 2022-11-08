// Copyright 2017 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package tool

import (
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strings"
)

// IsODMLFile returns true of the file has an odML header
func IsODMLFile(data []byte) bool {
	if !IsTextFile(data) {
		return false
	}
	return strings.Contains(string(data), "<odML version=")
}

// IsTextFile returns true if file content format is plain text or empty.
func IsTextFile(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	return strings.Contains(http.DetectContentType(data), "text/")
}

var RE_ANNEXPOINTERFILE = regexp.MustCompile(`^(/annex/objects/([A-Z][\-_0-9A-Za-z]+)(?:\n|\r|\z))`)

//reference: https://git-annex.branchable.com/internals/pointer_file/
func IsAnnexedFile(data []byte) bool {

	const ANNEXPOINTERFILE_MAXSIZE = 32 * 1024
	const ANNEXSNIFFSIZE = 512

	var dataLen = len(data)

	//The maximum size of a pointer file is 32 kb. If it is any longer, it is not considered to be a valid pointer file.
	if dataLen > ANNEXPOINTERFILE_MAXSIZE {
		return false
	}

	var sniffData []byte
	if !(dataLen < ANNEXSNIFFSIZE) {
		sniffData = data[:ANNEXSNIFFSIZE]
	} else {
		sniffData = data
	}

	//annex pointer file is a text file
	if strings.Contains(http.DetectContentType(sniffData), "text/") {

		//A pointer file starts with "/annex/objects/", which is followed by the key
		matchAnnexPointer := RE_ANNEXPOINTERFILE.FindStringSubmatch(string(sniffData))

		if len(matchAnnexPointer) > 0 {
			//var annexKey = matchAnnexPointer[2]

			//git-annex does support pointer files with additional text on subsequent lines.
			var hasAdditionalText = len(sniffData) > len(matchAnnexPointer[1]) || dataLen > ANNEXSNIFFSIZE

			if hasAdditionalText {
				//every such subsequent line must contain "/annex/" somewhere in it, and end with a newline.
				var extraLines = strings.SplitAfter(string(data), "\n")[1:]

				if extraLines[len(extraLines)-1] != "" {
					//if last line isn't empty, it means it was missing required newline character
					return false
				} else {
					for _, line := range extraLines[0 : len(extraLines)-1] {
						if !strings.Contains(line, "/annex/") {
							return false
						}
					}
				}
			}
			return true
		}
	}
	return false
}

func IsImageFile(data []byte) bool {
	return strings.Contains(http.DetectContentType(data), "image/")
}

func IsPDFFile(data []byte) bool {
	return strings.Contains(http.DetectContentType(data), "application/pdf")
}

func IsVideoFile(data []byte) bool {
	return strings.Contains(http.DetectContentType(data), "video/")
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

func humanateBytes(s uint64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := float64(s) / math.Pow(base, math.Floor(e))
	f := "%.0f"
	if val < 10 {
		f = "%.1f"
	}

	return fmt.Sprintf(f+" %s", val, suffix)
}

// FileSize calculates the file size and generate user-friendly string.
func FileSize(s int64) string {
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	return humanateBytes(uint64(s), 1024, sizes)
}
