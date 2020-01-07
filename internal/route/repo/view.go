// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"bytes"
	"fmt"
	gotemplate "html/template"
	"io/ioutil"
	"path"
	"strings"

	"github.com/go-macaron/captcha"
	"github.com/unknwon/paginater"
	log "gopkg.in/clog.v1"

	"github.com/G-Node/git-module"
	"github.com/G-Node/libgin/libgin/annex"

	"github.com/G-Node/gogs/internal/context"
	"github.com/G-Node/gogs/internal/db"
	"github.com/G-Node/gogs/internal/markup"
	"github.com/G-Node/gogs/internal/setting"
	"github.com/G-Node/gogs/internal/template"
	"github.com/G-Node/gogs/internal/template/highlight"
	"github.com/G-Node/gogs/internal/tool"
)

const (
	BARE     = "repo/bare"
	HOME     = "repo/home"
	WATCHERS = "repo/watchers"
	FORKS    = "repo/forks"
)

func renderDirectory(c *context.Context, treeLink string) {
	tree, err := c.Repo.Commit.SubTree(c.Repo.TreePath)
	if err != nil {
		c.NotFoundOrServerError("Repo.Commit.SubTree", git.IsErrNotExist, err)
		return
	}

	entries, err := tree.ListEntries()
	if err != nil {
		c.ServerError("ListEntries", err)
		return
	}
	entries.Sort()

	c.Data["Files"], err = entries.GetCommitsInfoWithCustomConcurrency(c.Repo.Commit, c.Repo.TreePath, setting.Repository.CommitsFetchConcurrency)
	if err != nil {
		c.ServerError("GetCommitsInfoWithCustomConcurrency", err)
		return
	}
	c.Data["HasDatacite"] = false
	var readmeFile *git.Blob
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() == "datacite.yml" {
			readDataciteFile(entry, c)
			break
		}
	}
	for _, entry := range entries {
		if entry.IsDir() || !markup.IsReadmeFile(entry.Name()) {
			continue
		}

		// TODO: collect all possible README files and show with priority.
		readmeFile = entry.Blob()
		break
	}

	if readmeFile != nil {
		c.Data["RawFileLink"] = ""
		c.Data["ReadmeInList"] = true
		c.Data["ReadmeExist"] = true

		dataRc, err := readmeFile.Data()
		if err != nil {
			c.ServerError("readmeFile.Data", err)
			return
		}

		buf := make([]byte, 1024)
		n, _ := dataRc.Read(buf)
		buf = buf[:n]

		// GIN mod: Replace existing buf and reader with annexed content buf and reader
		buf, dataRc, err = resolveAnnexedContent(c, buf, dataRc)
		if err != nil {
			return
		}

		isTextFile := tool.IsTextFile(buf)
		c.Data["IsTextFile"] = isTextFile
		c.Data["FileName"] = readmeFile.Name()
		if isTextFile {
			d, _ := ioutil.ReadAll(dataRc)
			buf = append(buf, d...)

			switch markup.Detect(readmeFile.Name()) {
			case markup.MARKDOWN:
				c.Data["IsMarkdown"] = true
				buf = markup.Markdown(buf, treeLink, c.Repo.Repository.ComposeMetas())
			case markup.ORG_MODE:
				c.Data["IsMarkdown"] = true
				buf = markup.OrgMode(buf, treeLink, c.Repo.Repository.ComposeMetas())
			case markup.IPYTHON_NOTEBOOK:
				c.Data["IsIPythonNotebook"] = true
				c.Data["RawFileLink"] = c.Repo.RepoLink + "/raw/" + path.Join(c.Repo.BranchName, c.Repo.TreePath, readmeFile.Name())
			default:
				buf = bytes.Replace(buf, []byte("\n"), []byte(`<br>`), -1)
			}
			c.Data["FileContent"] = string(buf)
		}
	}

	// Show latest commit info of repository in table header,
	// or of directory if not in root directory.
	latestCommit := c.Repo.Commit
	if len(c.Repo.TreePath) > 0 {
		latestCommit, err = c.Repo.Commit.GetCommitByPath(c.Repo.TreePath)
		if err != nil {
			c.ServerError("GetCommitByPath", err)
			return
		}
	}
	c.Data["LatestCommit"] = latestCommit
	c.Data["LatestCommitUser"] = db.ValidateCommitWithEmail(latestCommit)

	if c.Repo.CanEnableEditor() {
		c.Data["CanAddFile"] = true
		c.Data["CanUploadFile"] = setting.Repository.Upload.Enabled
	}
}

func renderFile(c *context.Context, entry *git.TreeEntry, treeLink, rawLink string, cpt *captcha.Captcha) {
	c.Data["IsViewFile"] = true

	blob := entry.Blob()
	dataRc, err := blob.Data()
	if err != nil {
		c.Handle(500, "Data", err)
		return
	}

	c.Data["FileSize"] = blob.Size()
	c.Data["FileName"] = blob.Name()
	c.Data["HighlightClass"] = highlight.FileNameToHighlightClass(blob.Name())
	c.Data["RawFileLink"] = rawLink + "/" + c.Repo.TreePath

	buf := make([]byte, 1024)
	n, _ := dataRc.Read(buf)
	buf = buf[:n]

	// GIN mod: Replace existing buf and reader with annexed content buf and reader
	buf, dataRc, err = resolveAnnexedContent(c, buf, dataRc)
	if err != nil {
		return
	}

	isTextFile := tool.IsTextFile(buf)
	c.Data["IsTextFile"] = isTextFile

	// Assume file is not editable first.
	if !isTextFile {
		c.Data["EditFileTooltip"] = c.Tr("repo.editor.cannot_edit_non_text_files")
	}

	canEnableEditor := c.Repo.CanEnableEditor()
	switch {
	case isTextFile:
		// GIN mod: Use c.Data["FileSize"] which is replaced by annexed content
		// size in resolveAnnexedContent() when necessary
		if c.Data["FileSize"].(int64) >= setting.UI.MaxDisplayFileSize {
			c.Data["IsFileTooLarge"] = true
			break
		}

		c.Data["ReadmeExist"] = markup.IsReadmeFile(blob.Name())

		d, _ := ioutil.ReadAll(dataRc)
		buf = append(buf, d...)

		switch markup.Detect(blob.Name()) {
		case markup.MARKDOWN:
			c.Data["IsMarkdown"] = true
			c.Data["FileContent"] = string(markup.Markdown(buf, path.Dir(treeLink), c.Repo.Repository.ComposeMetas()))
		case markup.ORG_MODE:
			c.Data["IsMarkdown"] = true
			c.Data["FileContent"] = string(markup.OrgMode(buf, path.Dir(treeLink), c.Repo.Repository.ComposeMetas()))
		case markup.IPYTHON_NOTEBOOK:
			c.Data["IsIPythonNotebook"] = true
			// GIN mod: JSON, YAML, and odML render support with jsTree
		case markup.JSON:
			c.Data["IsJSON"] = true
			c.Data["RawFileContent"] = string(buf)
			fallthrough
		case markup.YAML:
			c.Data["IsYAML"] = true
			c.Data["RawFileContent"] = string(buf)
			fallthrough
		case markup.XML:
			// pass XML down to ODML checker
			fallthrough
		case markup.ODML:
			if tool.IsODMLFile(buf) {
				c.Data["IsODML"] = true
				c.Data["ODML"] = string(markup.MarshalODML(buf))
			}
			fallthrough
		default:
			// Building code view blocks with line number on server side.
			var fileContent string
			if err, content := template.ToUTF8WithErr(buf); err != nil {
				if err != nil {
					log.Error(4, "ToUTF8WithErr: %s", err)
				}
				fileContent = string(buf)
			} else {
				fileContent = content
			}

			var output bytes.Buffer
			lines := strings.Split(fileContent, "\n")
			// Remove blank line at the end of file
			if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
				lines = lines[:len(lines)-1]
			}
			// > GIN
			if len(lines) > setting.UI.MaxLineHighlight {
				c.Data["HighlightClass"] = "nohighlight"
			}
			// < GIN
			for index, line := range lines {
				output.WriteString(fmt.Sprintf(`<li class="L%d" rel="L%d">%s</li>`, index+1, index+1, gotemplate.HTMLEscapeString(strings.TrimRight(line, "\r"))) + "\n")
			}
			c.Data["FileContent"] = gotemplate.HTML(output.String())

			output.Reset()
			for i := 0; i < len(lines); i++ {
				output.WriteString(fmt.Sprintf(`<span id="L%d">%d</span>`, i+1, i+1))
			}
			c.Data["LineNums"] = gotemplate.HTML(output.String())
		}

		isannex := tool.IsAnnexedFile(buf)
		if canEnableEditor && !isannex {
			c.Data["CanEditFile"] = true
			c.Data["EditFileTooltip"] = c.Tr("repo.editor.edit_this_file")
		} else if !c.Repo.IsViewBranch {
			c.Data["EditFileTooltip"] = c.Tr("repo.editor.must_be_on_a_branch")
		} else if !c.Repo.IsWriter() {
			c.Data["EditFileTooltip"] = c.Tr("repo.editor.fork_before_edit")
		}

	case tool.IsPDFFile(buf) && (c.Data["FileSize"].(int64) < setting.Repository.RawCaptchaMinFileSize*annex.MEGABYTE ||
		c.IsLogged):
		c.Data["IsPDFFile"] = true
	case tool.IsVideoFile(buf) && (c.Data["FileSize"].(int64) < setting.Repository.RawCaptchaMinFileSize*annex.MEGABYTE ||
		c.IsLogged):
		c.Data["IsVideoFile"] = true
	case tool.IsImageFile(buf) && (c.Data["FileSize"].(int64) < setting.Repository.RawCaptchaMinFileSize*annex.MEGABYTE ||
		c.IsLogged):
		c.Data["IsImageFile"] = true
	case tool.IsAnnexedFile(buf) && (c.Data["FileSize"].(int64) < setting.Repository.RawCaptchaMinFileSize*annex.MEGABYTE ||
		c.IsLogged):
		c.Data["IsAnnexedFile"] = true
	}

	if canEnableEditor {
		c.Data["CanDeleteFile"] = true
		c.Data["DeleteFileTooltip"] = c.Tr("repo.editor.delete_this_file")
	} else if !c.Repo.IsViewBranch {
		c.Data["DeleteFileTooltip"] = c.Tr("repo.editor.must_be_on_a_branch")
	} else if !c.Repo.IsWriter() {
		c.Data["DeleteFileTooltip"] = c.Tr("repo.editor.must_have_write_access")
	}
}

func setEditorconfigIfExists(c *context.Context) {
	ec, err := c.Repo.GetEditorconfig()
	if err != nil && !git.IsErrNotExist(err) {
		log.Trace("setEditorconfigIfExists.GetEditorconfig [%d]: %v", c.Repo.Repository.ID, err)
		return
	}
	c.Data["Editorconfig"] = ec
}

func Home(c *context.Context, cpt *captcha.Captcha) {
	c.Data["PageIsViewFiles"] = true

	if c.Repo.Repository.IsBare {
		c.HTML(200, BARE)
		return
	}

	title := c.Repo.Repository.Owner.Name + "/" + c.Repo.Repository.Name
	if len(c.Repo.Repository.Description) > 0 {
		title += ": " + c.Repo.Repository.Description
	}
	c.Data["Title"] = title
	if c.Repo.BranchName != c.Repo.Repository.DefaultBranch {
		c.Data["Title"] = title + " @ " + c.Repo.BranchName
	}
	c.Data["RequireHighlightJS"] = true

	branchLink := c.Repo.RepoLink + "/src/" + c.Repo.BranchName
	treeLink := branchLink
	rawLink := c.Repo.RepoLink + "/raw/" + c.Repo.BranchName

	isRootDir := false
	if len(c.Repo.TreePath) > 0 {
		treeLink += "/" + c.Repo.TreePath
	} else {
		isRootDir = true

		// Only show Git stats panel when view root directory
		var err error
		c.Repo.CommitsCount, err = c.Repo.Commit.CommitsCount()
		if err != nil {
			c.Handle(500, "CommitsCount", err)
			return
		}
		c.Data["CommitsCount"] = c.Repo.CommitsCount
	}
	c.Data["PageIsRepoHome"] = isRootDir

	// Get current entry user currently looking at.
	entry, err := c.Repo.Commit.GetTreeEntryByPath(c.Repo.TreePath)
	if err != nil {
		c.NotFoundOrServerError("Repo.Commit.GetTreeEntryByPath", git.IsErrNotExist, err)
		return
	}

	if entry.IsDir() {
		renderDirectory(c, treeLink)
	} else {
		renderFile(c, entry, treeLink, rawLink, cpt)
	}
	if c.Written() {
		return
	}

	setEditorconfigIfExists(c)
	if c.Written() {
		return
	}

	var treeNames []string
	paths := make([]string, 0, 5)
	if len(c.Repo.TreePath) > 0 {
		treeNames = strings.Split(c.Repo.TreePath, "/")
		for i := range treeNames {
			paths = append(paths, strings.Join(treeNames[:i+1], "/"))
		}

		c.Data["HasParentPath"] = true
		if len(paths)-2 >= 0 {
			c.Data["ParentPath"] = "/" + paths[len(paths)-2]
		}
	}

	c.Data["Paths"] = paths
	c.Data["TreeLink"] = treeLink
	c.Data["TreeNames"] = treeNames
	c.Data["BranchLink"] = branchLink
	c.HTML(200, HOME)
}

func RenderUserCards(c *context.Context, total int, getter func(page int) ([]*db.User, error), tpl string) {
	page := c.QueryInt("page")
	if page <= 0 {
		page = 1
	}
	pager := paginater.New(total, db.ItemsPerPage, page, 5)
	c.Data["Page"] = pager

	items, err := getter(pager.Current())
	if err != nil {
		c.Handle(500, "getter", err)
		return
	}
	c.Data["Cards"] = items

	c.HTML(200, tpl)
}

func Watchers(c *context.Context) {
	c.Data["Title"] = c.Tr("repo.watchers")
	c.Data["CardsTitle"] = c.Tr("repo.watchers")
	c.Data["PageIsWatchers"] = true
	RenderUserCards(c, c.Repo.Repository.NumWatches, c.Repo.Repository.GetWatchers, WATCHERS)
}

func Stars(c *context.Context) {
	c.Data["Title"] = c.Tr("repo.stargazers")
	c.Data["CardsTitle"] = c.Tr("repo.stargazers")
	c.Data["PageIsStargazers"] = true
	RenderUserCards(c, c.Repo.Repository.NumStars, c.Repo.Repository.GetStargazers, WATCHERS)
}

func Forks(c *context.Context) {
	c.Data["Title"] = c.Tr("repos.forks")

	forks, err := c.Repo.Repository.GetForks()
	if err != nil {
		c.Handle(500, "GetForks", err)
		return
	}

	for _, fork := range forks {
		if err = fork.GetOwner(); err != nil {
			c.Handle(500, "GetOwner", err)
			return
		}
	}
	c.Data["Forks"] = forks

	c.HTML(200, FORKS)
}
