// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package template

import (
	"container/list"
	"fmt"
	"html/template"
	"mime"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/json-iterator/go"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
	log "gopkg.in/clog.v1"
	"gopkg.in/editorconfig/editorconfig-core-go.v1"

	"github.com/G-Node/gogs/models"
	"github.com/G-Node/gogs/pkg/tool"
	"github.com/G-Node/gogs/pkg/markup"
	"github.com/G-Node/gogs/pkg/setting"
)

// TODO: only initialize map once and save to a local variable to reduce copies.
func NewFuncMap() []template.FuncMap {
	return []template.FuncMap{map[string]interface{}{
		"GoVer": func() string {
			return strings.Title(runtime.Version())
		},
		"UseHTTPS": func() bool {
			return strings.HasPrefix(setting.AppURL, "https")
		},
		"AppName": func() string {
			return setting.AppName
		},
		"AppSubURL": func() string {
			return setting.AppSubURL
		},
		"AppURL": func() string {
			return setting.AppURL
		},
		"AppVer": func() string {
			return setting.AppVer
		},
		"AppDomain": func() string {
			return setting.Domain
		},
		"DisableGravatar": func() bool {
			return setting.DisableGravatar
		},
		"ShowFooterTemplateLoadTime": func() bool {
			return setting.ShowFooterTemplateLoadTime
		},
		"LoadTimes": func(startTime time.Time) string {
			return fmt.Sprint(time.Since(startTime).Nanoseconds()/1e6) + "ms"
		},
		"AvatarLink":       tool.AvatarLink,
		"AppendAvatarSize": tool.AppendAvatarSize,
		"Safe":             Safe,
		"Sanitize":         bluemonday.UGCPolicy().Sanitize,
		"Str2HTML":         Str2HTML,
		"NewLine2br":       NewLine2br,
		"TimeSince":        tool.TimeSince,
		"RawTimeSince":     tool.RawTimeSince,
		"FileSize":         tool.FileSize,
		"Subtract":         tool.Subtract,
		"Add": func(a, b int) int {
			return a + b
		},
		"ActionIcon": ActionIcon,
		"DateFmtLong": func(t time.Time) string {
			return t.Format(time.RFC1123Z)
		},
		"DateFmtShort": func(t time.Time) string {
			return t.Format("Jan 02, 2006")
		},
		"List": List,
		"SubStr": func(str string, start, length int) string {
			if len(str) == 0 {
				return ""
			}
			end := start + length
			if length == -1 {
				end = len(str)
			}
			if len(str) < end {
				return str
			}
			return str[start:end]
		},
		"Join":              strings.Join,
		"EllipsisString":    tool.EllipsisString,
		"DiffTypeToStr":     DiffTypeToStr,
		"DiffLineTypeToStr": DiffLineTypeToStr,
		"Sha1":              Sha1,
		"ShortSHA1":         tool.ShortSHA1,
		"MD5":               tool.MD5,
		"ActionContent2Commits": ActionContent2Commits,
		"EscapePound":           EscapePound,
		"RenderCommitMessage":   RenderCommitMessage,
		"ThemeColorMetaTag": func() string {
			return setting.UI.ThemeColorMetaTag
		},
		"FilenameIsImage": func(filename string) bool {
			mimeType := mime.TypeByExtension(filepath.Ext(filename))
			return strings.HasPrefix(mimeType, "image/")
		},
		"TabSizeClass": func(ec *editorconfig.Editorconfig, filename string) string {
			if ec != nil {
				def := ec.GetDefinitionForFilename(filename)
				if def.TabWidth > 0 {
					return fmt.Sprintf("tab-size-%d", def.TabWidth)
				}
			}
			return "tab-size-8"
		},
	}}
}

func Safe(raw string) template.HTML {
	return template.HTML(raw)
}

func Str2HTML(raw string) template.HTML {
	return template.HTML(markup.Sanitize(raw))
}

// NewLine2br simply replaces "\n" to "<br>".
func NewLine2br(raw string) string {
	return strings.Replace(raw, "\n", "<br>", -1)
}

func List(l *list.List) chan interface{} {
	e := l.Front()
	c := make(chan interface{})
	go func() {
		for e != nil {
			c <- e.Value
			e = e.Next()
		}
		close(c)
	}()
	return c
}

func Sha1(str string) string {
	return tool.SHA1(str)
}

func ToUTF8WithErr(content []byte) (error, string) {
	charsetLabel, err := tool.DetectEncoding(content)
	if err != nil {
		return err, ""
	} else if charsetLabel == "UTF-8" {
		return nil, string(content)
	}

	encoding, _ := charset.Lookup(charsetLabel)
	if encoding == nil {
		return fmt.Errorf("Unknown encoding: %s", charsetLabel), string(content)
	}

	// If there is an error, we concatenate the nicely decoded part and the
	// original left over. This way we won't loose data.
	result, n, err := transform.String(encoding.NewDecoder(), string(content))
	if err != nil {
		result = result + string(content[n:])
	}

	return err, result
}

// FIXME: Unused function
func ToUTF8(content string) string {
	_, res := ToUTF8WithErr([]byte(content))
	return res
}

// Replaces all prefixes 'old' in 's' with 'new'.
// FIXME: Unused function
func ReplaceLeft(s, old, new string) string {
	old_len, new_len, i, n := len(old), len(new), 0, 0
	for ; i < len(s) && strings.HasPrefix(s[i:], old); n += 1 {
		i += old_len
	}

	// simple optimization
	if n == 0 {
		return s
	}

	// allocating space for the new string
	newLen := n*new_len + len(s[i:])
	replacement := make([]byte, newLen, newLen)

	j := 0
	for ; j < n*new_len; j += new_len {
		copy(replacement[j:j+new_len], new)
	}

	copy(replacement[j:], s[i:])
	return string(replacement)
}

// RenderCommitMessage renders commit message with special links.
func RenderCommitMessage(full bool, msg, urlPrefix string, metas map[string]string) string {
	cleanMsg := template.HTMLEscapeString(msg)
	fullMessage := string(markup.RenderIssueIndexPattern([]byte(cleanMsg), urlPrefix, metas))
	msgLines := strings.Split(strings.TrimSpace(fullMessage), "\n")
	numLines := len(msgLines)
	if numLines == 0 {
		return ""
	} else if !full {
		return msgLines[0]
	} else if numLines == 1 || (numLines >= 2 && len(msgLines[1]) == 0) {
		// First line is a header, standalone or followed by empty line
		header := fmt.Sprintf("<h3>%s</h3>", msgLines[0])
		if numLines >= 2 {
			fullMessage = header + fmt.Sprintf("\n<pre>%s</pre>", strings.Join(msgLines[2:], "\n"))
		} else {
			fullMessage = header
		}
	} else {
		// Non-standard git message, there is no header line
		fullMessage = fmt.Sprintf("<h4>%s</h4>", strings.Join(msgLines, "<br>"))
	}
	return fullMessage
}

type Actioner interface {
	GetOpType() int
	GetActUserName() string
	GetRepoUserName() string
	GetRepoName() string
	GetRepoPath() string
	GetRepoLink() string
	GetBranch() string
	GetContent() string
	GetCreate() time.Time
	GetIssueInfos() []string
}

// ActionIcon accepts a int that represents action operation type
// and returns a icon class name.
func ActionIcon(opType int) string {
	switch opType {
	case 1, 8: // Create and transfer repository
		return "repo"
	case 5: // Commit repository
		return "git-commit"
	case 6: // Create issue
		return "issue-opened"
	case 7: // New pull request
		return "git-pull-request"
	case 9: // Push tag
		return "tag"
	case 10: // Comment issue
		return "comment-discussion"
	case 11: // Merge pull request
		return "git-merge"
	case 12, 14: // Close issue or pull request
		return "issue-closed"
	case 13, 15: // Reopen issue or pull request
		return "issue-reopened"
	case 16: // Create branch
		return "git-branch"
	case 17, 18: // Delete branch or tag
		return "alert"
	case 19: // Fork a repository
		return "repo-forked"
	case 20, 21, 22: // Mirror sync
		return "repo-clone"
	default:
		return "invalid type"
	}
}

func ActionContent2Commits(act Actioner) *models.PushCommits {
	push := models.NewPushCommits()
	if err := jsoniter.Unmarshal([]byte(act.GetContent()), push); err != nil {
		log.Error(4, "Unmarshal:\n%s\nERROR: %v", act.GetContent(), err)
	}
	return push
}

func EscapePound(str string) string {
	return strings.NewReplacer("%", "%25", "#", "%23", " ", "%20", "?", "%3F").Replace(str)
}

func DiffTypeToStr(diffType int) string {
	diffTypes := map[int]string{
		1: "add", 2: "modify", 3: "del", 4: "rename",
	}
	return diffTypes[diffType]
}

func DiffLineTypeToStr(diffType int) string {
	switch diffType {
	case 2:
		return "add"
	case 3:
		return "del"
	case 4:
		return "tag"
	}
	return "same"
}
