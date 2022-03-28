// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	log "unknwon.dev/clog/v2"

	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/db"
	"github.com/NII-DG/gogs/internal/db/errors"
	"github.com/NII-DG/gogs/internal/form"
	"github.com/NII-DG/gogs/internal/gitutil"
	"github.com/NII-DG/gogs/internal/markup"
	"github.com/NII-DG/gogs/internal/pathutil"
	"github.com/NII-DG/gogs/internal/template"
	"github.com/NII-DG/gogs/internal/tool"
)

const (
	tmplEditorEdit        = "repo/editor/edit"
	tmplEditorDiffPreview = "repo/editor/diff_preview"
	tmplEditorDelete      = "repo/editor/delete"
	tmplEditorUpload      = "repo/editor/upload"
)

// getParentTreeFields returns list of parent tree names and corresponding tree paths
// based on given tree path.
func getParentTreeFields(treePath string) (treeNames []string, treePaths []string) {
	if len(treePath) == 0 {
		return treeNames, treePaths
	}

	treeNames = strings.Split(treePath, "/")
	treePaths = make([]string, len(treeNames))
	for i := range treeNames {
		treePaths[i] = strings.Join(treeNames[:i+1], "/")
	}
	return treeNames, treePaths
}

func editFile(c *context.Context, isNewFile bool) {
	c.PageIs("Edit")
	c.RequireHighlightJS()
	c.RequireSimpleMDE()
	c.Data["IsNewFile"] = isNewFile

	treeNames, treePaths := getParentTreeFields(c.Repo.TreePath)

	if !isNewFile {
		entry, err := c.Repo.Commit.TreeEntry(c.Repo.TreePath)
		if err != nil {
			c.NotFoundOrError(gitutil.NewError(err), "get tree entry")
			return
		}

		// No way to edit a directory online.
		if entry.IsTree() {
			c.NotFound()
			return
		}

		blob := entry.Blob()
		buf, err := blob.Bytes()
		if err != nil {
			c.Error(err, "get blob data")
			return
		}

		if blob.Name() == "dmp.json" {
			// DMPであればJSONスキーマの情報も取得する
			dmpSchema := &struct{ Schema string }{}
			if err := json.Unmarshal(buf, &dmpSchema); err != nil {
				log.Error("DMP data can't be unmarshalled: %v", err)
				c.Data["HasDmpJson"] = false
			} else {
				schemaUrl := getTemplateUrl()

				var d dmpUtil

				err = d.FetchDmpSchema(c, schemaUrl+"dmp/json_schema/schema_dmp_"+dmpSchema.Schema)
				if err != nil {
					log.Error("failed fetching DMP template: %v", err)
				}
			}
		}

		c.Data["FileSize"] = blob.Size()
		c.Data["FileName"] = blob.Name()
		c.Data["IsJSON"] = markup.IsJSON(blob.Name())
		c.Data["IsYAML"] = markup.IsYAML(blob.Name())

		// Only text file are editable online.
		if !tool.IsTextFile(buf) {
			c.NotFound()
			return
		}

		c.Data["IsODML"] = tool.IsODMLFile(buf)

		if err, content := template.ToUTF8WithErr(buf); err != nil {
			if err != nil {
				log.Error("Failed to convert encoding to UTF-8: %v", err)
			}
			c.Data["FileContent"] = string(buf)
		} else {
			c.Data["FileContent"] = content
		}
		c.Data["commit_summary"] = "実施事項: " + blob.Name() + "を編集"
	} else {
		treeNames = append(treeNames, "") // Append empty string to allow user name the new file.
		c.Data["commit_summary"] = "実施事項: ファイルを新規作成"
	}

	c.Data["ParentTreePath"] = path.Dir(c.Repo.TreePath)
	c.Data["TreeNames"] = treeNames
	c.Data["TreePaths"] = treePaths
	c.Data["BranchLink"] = c.Repo.RepoLink + "/src/" + c.Repo.BranchName

	c.Data["commit_message"] = ""
	c.Data["commit_choice"] = "direct"
	c.Data["new_branch_name"] = ""
	c.Data["last_commit"] = c.Repo.Commit.ID
	c.Data["MarkdownFileExts"] = strings.Join(conf.Markdown.FileExtensions, ",")
	c.Data["LineWrapExtensions"] = strings.Join(conf.Repository.Editor.LineWrapExtensions, ",")
	c.Data["PreviewableFileModes"] = strings.Join(conf.Repository.Editor.PreviewableFileModes, ",")
	c.Data["EditorconfigURLPrefix"] = fmt.Sprintf("%s/api/v1/repos/%s/editorconfig/", conf.Server.Subpath, c.Repo.Repository.FullName())

	c.Success(tmplEditorEdit)
}

func EditFile(c *context.Context) {
	editFile(c, false)
}

func NewFile(c *context.Context) {
	editFile(c, true)
}

func editFilePost(c *context.Context, f form.EditRepoFile, isNewFile bool) {
	c.PageIs("Edit")
	c.RequireHighlightJS()
	c.RequireSimpleMDE()
	c.Data["IsNewFile"] = isNewFile

	oldBranchName := c.Repo.BranchName
	branchName := oldBranchName
	oldTreePath := c.Repo.TreePath
	lastCommit := f.LastCommit
	f.LastCommit = c.Repo.Commit.ID.String()

	if f.IsNewBrnach() {
		branchName = f.NewBranchName
	}

	f.TreePath = pathutil.Clean(f.TreePath)
	treeNames, treePaths := getParentTreeFields(f.TreePath)

	c.Data["ParentTreePath"] = path.Dir(c.Repo.TreePath)
	c.Data["TreePath"] = f.TreePath
	c.Data["TreeNames"] = treeNames
	c.Data["TreePaths"] = treePaths
	c.Data["BranchLink"] = c.Repo.RepoLink + "/src/" + branchName
	c.Data["FileContent"] = f.Content
	c.Data["commit_summary"] = f.CommitSummary
	c.Data["commit_message"] = f.CommitMessage
	c.Data["commit_choice"] = f.CommitChoice
	c.Data["new_branch_name"] = branchName
	c.Data["last_commit"] = f.LastCommit
	c.Data["MarkdownFileExts"] = strings.Join(conf.Markdown.FileExtensions, ",")
	c.Data["LineWrapExtensions"] = strings.Join(conf.Repository.Editor.LineWrapExtensions, ",")
	c.Data["PreviewableFileModes"] = strings.Join(conf.Repository.Editor.PreviewableFileModes, ",")

	if c.HasError() {
		c.Success(tmplEditorEdit)
		return
	}

	if len(f.TreePath) == 0 {
		c.FormErr("TreePath")
		c.RenderWithErr(c.Tr("repo.editor.filename_cannot_be_empty"), tmplEditorEdit, &f)
		return
	}

	if oldBranchName != branchName {
		if _, err := c.Repo.Repository.GetBranch(branchName); err == nil {
			c.FormErr("NewBranchName")
			c.RenderWithErr(c.Tr("repo.editor.branch_already_exists", branchName), tmplEditorEdit, &f)
			return
		}
	}

	var newTreePath string
	for index, part := range treeNames {
		newTreePath = path.Join(newTreePath, part)
		entry, err := c.Repo.Commit.TreeEntry(newTreePath)
		if err != nil {
			if gitutil.IsErrRevisionNotExist(err) {
				// Means there is no item with that name, so we're good
				break
			}

			c.Error(err, "get tree entry")
			return
		}
		if index != len(treeNames)-1 {
			if !entry.IsTree() {
				c.FormErr("TreePath")
				c.RenderWithErr(c.Tr("repo.editor.directory_is_a_file", part), tmplEditorEdit, &f)
				return
			}
		} else {
			if entry.IsSymlink() {
				c.FormErr("TreePath")
				c.RenderWithErr(c.Tr("repo.editor.file_is_a_symlink", part), tmplEditorEdit, &f)
				return
			} else if entry.IsTree() {
				c.FormErr("TreePath")
				c.RenderWithErr(c.Tr("repo.editor.filename_is_a_directory", part), tmplEditorEdit, &f)
				return
			}
		}
	}

	if !isNewFile {
		_, err := c.Repo.Commit.TreeEntry(oldTreePath)
		if err != nil {
			if gitutil.IsErrRevisionNotExist(err) {
				c.FormErr("TreePath")
				c.RenderWithErr(c.Tr("repo.editor.file_editing_no_longer_exists", oldTreePath), tmplEditorEdit, &f)
			} else {
				c.Error(err, "get tree entry")
			}
			return
		}
		if lastCommit != c.Repo.CommitID {
			files, err := c.Repo.Commit.FilesChangedAfter(lastCommit)
			if err != nil {
				c.Error(err, "get changed files")
				return
			}

			for _, file := range files {
				if file == f.TreePath {
					c.RenderWithErr(c.Tr("repo.editor.file_changed_while_editing", c.Repo.RepoLink+"/compare/"+lastCommit+"..."+c.Repo.CommitID), tmplEditorEdit, &f)
					return
				}
			}
		}
	}

	if oldTreePath != f.TreePath {
		// We have a new filename (rename or completely new file) so we need to make sure it doesn't already exist, can't clobber.
		entry, err := c.Repo.Commit.TreeEntry(f.TreePath)
		if err != nil {
			if !gitutil.IsErrRevisionNotExist(err) {
				c.Error(err, "get tree entry")
				return
			}
		}
		if entry != nil {
			c.FormErr("TreePath")
			c.RenderWithErr(c.Tr("repo.editor.file_already_exists", f.TreePath), tmplEditorEdit, &f)
			return
		}
	}

	message := strings.TrimSpace(f.CommitSummary)
	if len(message) == 0 {
		if isNewFile {
			message = c.Tr("repo.editor.add", f.TreePath)
		} else {
			message = c.Tr("repo.editor.update", f.TreePath)
		}
	}

	f.CommitMessage = strings.TrimSpace(f.CommitMessage)
	if len(f.CommitMessage) > 0 {
		message += "\n\n" + f.CommitMessage
	}

	if err := c.Repo.Repository.UpdateRepoFile(c.User, db.UpdateRepoFileOptions{
		LastCommitID: lastCommit,
		OldBranch:    oldBranchName,
		NewBranch:    branchName,
		OldTreeName:  oldTreePath,
		NewTreeName:  f.TreePath,
		Message:      message,
		Content:      strings.Replace(f.Content, "\r", "", -1),
		IsNewFile:    isNewFile,
	}); err != nil {
		log.Error("Failed to update repo file: %v", err)
		c.FormErr("TreePath")
		c.RenderWithErr(c.Tr("repo.editor.fail_to_update_file", f.TreePath, errors.InternalServerError), tmplEditorEdit, &f)
		return
	}

	if f.IsNewBrnach() && c.Repo.PullRequest.Allowed {
		c.Redirect(c.Repo.PullRequestURL(oldBranchName, f.NewBranchName))
	} else {
		c.Redirect(c.Repo.RepoLink + "/src/" + branchName + "/" + f.TreePath)
	}
}

func EditFilePost(c *context.Context, f form.EditRepoFile) {
	editFilePost(c, f, false)
}

func NewFilePost(c *context.Context, f form.EditRepoFile) {
	editFilePost(c, f, true)
}

func DiffPreviewPost(c *context.Context, f form.EditPreviewDiff) {
	treePath := c.Repo.TreePath

	entry, err := c.Repo.Commit.TreeEntry(treePath)
	if err != nil {
		c.Error(err, "get tree entry")
		return
	} else if entry.IsTree() {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	diff, err := c.Repo.Repository.GetDiffPreview(c.Repo.BranchName, treePath, f.Content)
	if err != nil {
		c.Error(err, "get diff preview")
		return
	}

	if diff.NumFiles() == 0 {
		c.PlainText(http.StatusOK, c.Tr("repo.editor.no_changes_to_show"))
		return
	}
	c.Data["File"] = diff.Files[0]

	c.Success(tmplEditorDiffPreview)
}

func DeleteFile(c *context.Context) {
	c.PageIs("Delete")

	entry, err := c.Repo.Commit.TreeEntry(c.Repo.TreePath)
	if err != nil {
		c.NotFoundOrError(gitutil.NewError(err), "get tree entry")
		return
	}

	// No way to edit a directory online.
	if entry.IsTree() {
		c.NotFound()
		return
	}
	blob := entry.Blob()

	c.Data["BranchLink"] = c.Repo.RepoLink + "/src/" + c.Repo.BranchName
	c.Data["TreePath"] = c.Repo.TreePath
	c.Data["commit_summary"] = "実施事項: " + blob.Name() + "を削除"
	c.Data["commit_message"] = ""
	c.Data["commit_choice"] = "direct"
	c.Data["new_branch_name"] = ""
	c.Success(tmplEditorDelete)
}

func DeleteFilePost(c *context.Context, f form.DeleteRepoFile) {
	c.PageIs("Delete")
	c.Data["BranchLink"] = c.Repo.RepoLink + "/src/" + c.Repo.BranchName

	c.Repo.TreePath = pathutil.Clean(c.Repo.TreePath)
	c.Data["TreePath"] = c.Repo.TreePath

	oldBranchName := c.Repo.BranchName
	branchName := oldBranchName

	if f.IsNewBrnach() {
		branchName = f.NewBranchName
	}
	c.Data["commit_summary"] = f.CommitSummary
	c.Data["commit_message"] = f.CommitMessage
	c.Data["commit_choice"] = f.CommitChoice
	c.Data["new_branch_name"] = branchName

	if c.HasError() {
		c.Success(tmplEditorDelete)
		return
	}

	if oldBranchName != branchName {
		if _, err := c.Repo.Repository.GetBranch(branchName); err == nil {
			c.FormErr("NewBranchName")
			c.RenderWithErr(c.Tr("repo.editor.branch_already_exists", branchName), tmplEditorDelete, &f)
			return
		}
	}

	message := strings.TrimSpace(f.CommitSummary)
	if len(message) == 0 {
		message = c.Tr("repo.editor.delete", c.Repo.TreePath)
	}

	f.CommitMessage = strings.TrimSpace(f.CommitMessage)
	if len(f.CommitMessage) > 0 {
		message += "\n\n" + f.CommitMessage
	}

	if err := c.Repo.Repository.DeleteRepoFile(c.User, db.DeleteRepoFileOptions{
		LastCommitID: c.Repo.CommitID,
		OldBranch:    oldBranchName,
		NewBranch:    branchName,
		TreePath:     c.Repo.TreePath,
		Message:      message,
	}); err != nil {
		log.Error("Failed to delete repo file: %v", err)
		c.RenderWithErr(c.Tr("repo.editor.fail_to_delete_file", c.Repo.TreePath, errors.InternalServerError), tmplEditorDelete, &f)
		return
	}

	if f.IsNewBrnach() && c.Repo.PullRequest.Allowed {
		c.Redirect(c.Repo.PullRequestURL(oldBranchName, f.NewBranchName))
	} else {
		c.Flash.Success(c.Tr("repo.editor.file_delete_success", c.Repo.TreePath))
		c.Redirect(c.Repo.RepoLink + "/src/" + branchName)
	}
}

func renderUploadSettings(c *context.Context) {
	c.RequireDropzone()
	c.Data["UploadAllowedTypes"] = strings.Join(conf.Repository.Upload.AllowedTypes, ",")
	c.Data["UploadMaxSize"] = conf.Repository.Upload.FileMaxSize
	c.Data["UploadMaxFiles"] = conf.Repository.Upload.MaxFiles
}

func UploadFile(c *context.Context) {
	c.PageIs("Upload")
	renderUploadSettings(c)

	treeNames, treePaths := getParentTreeFields(c.Repo.TreePath)
	if len(treeNames) == 0 {
		// We must at least have one element for user to input.
		treeNames = []string{""}
	}

	c.Data["TreeNames"] = treeNames
	c.Data["TreePaths"] = treePaths
	c.Data["BranchLink"] = c.Repo.RepoLink + "/src/" + c.Repo.BranchName
	c.Data["commit_summary"] = "実施事項: ファイルをアップロード"
	c.Data["commit_message"] = ""
	c.Data["commit_choice"] = "direct"
	c.Data["new_branch_name"] = ""
	c.Success(tmplEditorUpload)
}

func UploadFilePost(c *context.Context, f form.UploadRepoFile) {
	c.PageIs("Upload")
	renderUploadSettings(c)

	oldBranchName := c.Repo.BranchName
	branchName := oldBranchName

	if f.IsNewBrnach() {
		branchName = f.NewBranchName
	}

	f.TreePath = pathutil.Clean(f.TreePath)
	treeNames, treePaths := getParentTreeFields(f.TreePath)
	if len(treeNames) == 0 {
		// We must at least have one element for user to input.
		treeNames = []string{""}
	}

	c.Data["TreePath"] = f.TreePath
	c.Data["TreeNames"] = treeNames
	c.Data["TreePaths"] = treePaths
	c.Data["BranchLink"] = c.Repo.RepoLink + "/src/" + branchName
	c.Data["commit_summary"] = f.CommitSummary
	c.Data["commit_message"] = f.CommitMessage
	c.Data["commit_choice"] = f.CommitChoice
	c.Data["new_branch_name"] = branchName

	if c.HasError() {
		c.Success(tmplEditorUpload)
		return
	}

	if oldBranchName != branchName {
		if _, err := c.Repo.Repository.GetBranch(branchName); err == nil {
			c.FormErr("NewBranchName")
			c.RenderWithErr(c.Tr("repo.editor.branch_already_exists", branchName), tmplEditorUpload, &f)
			return
		}
	}

	var newTreePath string
	for _, part := range treeNames {
		newTreePath = path.Join(newTreePath, part)
		entry, err := c.Repo.Commit.TreeEntry(newTreePath)
		if err != nil {
			if gitutil.IsErrRevisionNotExist(err) {
				// Means there is no item with that name, so we're good
				break
			}

			c.Error(err, "get tree entry")
			return
		}

		// User can only upload files to a directory.
		if !entry.IsTree() {
			c.FormErr("TreePath")
			c.RenderWithErr(c.Tr("repo.editor.directory_is_a_file", part), tmplEditorUpload, &f)
			return
		}
	}

	message := strings.TrimSpace(f.CommitSummary)
	if len(message) == 0 {
		message = c.Tr("repo.editor.upload_files_to_dir", f.TreePath)
	}

	f.CommitMessage = strings.TrimSpace(f.CommitMessage)
	if len(f.CommitMessage) > 0 {
		message += "\n\n" + f.CommitMessage
	}

	if err := c.Repo.Repository.UploadRepoFiles(c.User, db.UploadRepoFileOptions{
		LastCommitID: c.Repo.CommitID,
		OldBranch:    oldBranchName,
		NewBranch:    branchName,
		TreePath:     f.TreePath,
		Message:      message,
		Files:        f.Files,
	}); err != nil {
		log.Error("Failed to upload files: %v", err)
		c.FormErr("TreePath")
		c.RenderWithErr(c.Tr("repo.editor.unable_to_upload_files", f.TreePath, errors.InternalServerError), tmplEditorUpload, &f)
		return
	}

	if f.IsNewBrnach() && c.Repo.PullRequest.Allowed {
		c.Redirect(c.Repo.PullRequestURL(oldBranchName, f.NewBranchName))
	} else {
		c.Redirect(c.Repo.RepoLink + "/src/" + branchName + "/" + f.TreePath)
	}
}

func UploadFileToServer(c *context.Context) {
	file, header, err := c.Req.FormFile("file")
	fvalue := c.Req.Form
	fp := filepath.Dir(fvalue.Get("full_path"))
	log.Info("full_path: %s", fp)
	if err != nil {
		c.Error(err, "get file")
		return
	}
	defer file.Close()

	buf := make([]byte, 1024)
	n, _ := file.Read(buf)
	if n > 0 {
		buf = buf[:n]
	}
	fileType := http.DetectContentType(buf)

	if len(conf.Repository.Upload.AllowedTypes) > 0 {
		allowed := false
		for _, t := range conf.Repository.Upload.AllowedTypes {
			t := strings.Trim(t, " ")
			if t == "*/*" || t == fileType {
				allowed = true
				break
			}
		}

		if !allowed {
			c.PlainText(http.StatusBadRequest, ErrFileTypeForbidden.Error())
			return
		}
	}

	upload, err := db.NewUpload(filepath.Join(fp, header.Filename), buf, file)
	if err != nil {
		c.Error(err, "new upload")
		return
	}

	log.Trace("New file uploaded by user[%d]: %s", c.UserID(), upload.UUID)
	c.JSONSuccess(map[string]string{
		"uuid": upload.UUID,
	})
}

func RemoveUploadFileFromServer(c *context.Context, f form.RemoveUploadFile) {
	if len(f.File) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	if err := db.DeleteUploadByUUID(f.File); err != nil {
		c.Error(err, "delete upload by UUID")
		return
	}

	log.Trace("Upload file removed: %s", f.File)
	c.Status(http.StatusNoContent)
}

func CreateDmp(c context.AbstructContext) {
	var f repoUtil
	var d dmpUtil
	createDmp(c, f, d)
}

// CreateDmp is RCOS specific code
func createDmp(c context.AbstructContext, f AbstructRepoUtil, d AbstructDmpUtil) {
	schema := c.QueryEscape("schema")
	schemaUrl := getTemplateUrl() + "dmp/"
	treeNames, treePaths := getParentTreeFields(c.GetRepo().GetTreePath())

	c.PageIs("Edit")
	c.RequireHighlightJS()
	c.RequireSimpleMDE()

	// data binding for "Add DMP" pulldown at DMP editing page
	// (The pulldown on the repository top page is binded in repo.renderDirectory.)
	err := d.BidingDmpSchemaList(c, schemaUrl+"orgs")
	if err != nil {
		log.Error("%v", err)
		return
	}
	err = d.FetchDmpSchema(c, schemaUrl+"json_schema/schema_dmp_"+schema)
	if err != nil {
		log.Error("%v", err)
		return
	}

	srcBasic, err := f.FetchContentsOnGithub(schemaUrl + "basic")
	if err != nil {
		log.Error("%v", err)
		return
	}
	decodedBasicSchema, err := f.DecodeBlobContent(srcBasic)
	if err != nil {
		log.Error("%v", err)
		return
	}

	srcOrg, err := f.FetchContentsOnGithub(schemaUrl + "orgs/" + schema)
	if err != nil {
		log.Error("%v", err)
		return
	}
	decodedOrgSchema, err := f.DecodeBlobContent(srcOrg)
	if err != nil {
		log.Error("%v", err)
		return
	}

	combinedDmp := decodedBasicSchema + decodedOrgSchema

	c.CallData()["IsYAML"] = false
	c.CallData()["IsJSON"] = true
	c.CallData()["IsDmpJson"] = true

	c.CallData()["FileContent"] = combinedDmp
	c.CallData()["ParentTreePath"] = path.Dir(c.GetRepo().GetTreePath())
	c.CallData()["TreeNames"] = treeNames
	c.CallData()["TreePaths"] = treePaths
	c.CallData()["BranchLink"] = c.GetRepo().GetRepoLink() + "/src/" + c.GetRepo().GetBranchName()
	c.CallData()["commit_summary"] = "実施事項: dmp.jsonを新規作成"
	c.CallData()["commit_message"] = ""
	c.CallData()["commit_choice"] = "direct"
	c.CallData()["new_branch_name"] = ""
	c.CallData()["last_commit"] = c.GetRepo().GetCommitId()
	c.CallData()["MarkdownFileExts"] = strings.Join(conf.Markdown.FileExtensions, ",")
	c.CallData()["LineWrapExtensions"] = strings.Join(conf.Repository.Editor.LineWrapExtensions, ",")
	c.CallData()["PreviewableFileModes"] = strings.Join(conf.Repository.Editor.PreviewableFileModes, ",")
	c.CallData()["EditorconfigURLPrefix"] = fmt.Sprintf("%s/api/v1/repos/%s/editorconfig/", conf.Server.Subpath, c.GetRepo().GetDbRepo().FullName())

	c.Success(tmplEditorEdit)
}

type AbstructDmpUtil interface {
	FetchDmpSchema(c context.AbstructContext, blobPath string) error
	BidingDmpSchemaList(c context.AbstructContext, treePath string) error
}

// dmpUtil is an alias for utility functions related to the manipulation of DMP information.
// For effective unit test execution, the above DmpUtil interface must be satisfied.
type dmpUtil func()

func (d dmpUtil) FetchDmpSchema(c context.AbstructContext, blobPath string) error {
	var f repoUtil
	return d.fetchDmpSchema(c, f, blobPath)
}

func (d dmpUtil) BidingDmpSchemaList(c context.AbstructContext, treePath string) error {
	var f repoUtil
	return d.bidingDmpSchemaList(c, f, treePath)
}

// fetchDmpSchema is RCOS specific code.
// This function fetch&bind JSON Schema of DMP for validation.
func (d dmpUtil) fetchDmpSchema(c context.AbstructContext, f AbstructRepoUtil, blobPath string) error {
	src, err := f.FetchContentsOnGithub(blobPath)
	if err != nil {
		return err
	}

	decodedScheme, err := f.DecodeBlobContent(src)
	if err != nil {
		return err
	}

	c.CallData()["IsDmpJson"] = true
	c.CallData()["Schema"] = decodedScheme
	return nil
}

// bidingDmpSchemaList is RCOS specific code.
// This function binds DMP organization list.
func (d dmpUtil) bidingDmpSchemaList(c context.AbstructContext, f AbstructRepoUtil, treePath string) error {
	contents, err := f.FetchContentsOnGithub(treePath)
	if err != nil {
		return err
	}

	var orgsInfo interface{}
	err = json.Unmarshal(contents, &orgsInfo)
	if err != nil {
		return err
	}

	// create organization list
	orgs := orgsInfo.([]interface{})
	var schemaList []string
	for i := range orgs {
		org := orgs[i].(map[string]interface{})["name"]
		schemaList = append(schemaList, org.(string))
	}

	c.CallData()["SchemaList"] = schemaList
	return nil
}

// getTemplateUrl is RCOS specific code.
// This is a helper function that returns a base URL
// for retrieving DMP templates, etc. from GitHub.
func getTemplateUrl() string {
	return "https://api.github.com/repos/NII-DG/maDMP-template/contents/"
}
