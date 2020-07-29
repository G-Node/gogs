package repo

import (
	"fmt"

	"github.com/gogs/git-module"

	"gogs.io/gogs/internal/context"
)

type repoGitTree struct {
	Sha  string              `json:"sha"`
	URL  string              `json:"url"`
	Tree []*repoGitTreeEntry `json:"tree"`
}

type repoGitTreeEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	Size int64  `json:"size"`
	Sha  string `json:"sha"`
	URL  string `json:"url"`
}

func GetRepoGitTree(c *context.APIContext) {
	gitTree, err := c.Repo.GitRepo.GetTree(c.Params(":sha"))
	if err != nil {
		c.NotFoundOrServerError("GetRepoGitTree", git.IsErrNotExist, err)
		return
	}
	entries, err := gitTree.ListEntries()
	if err != nil {
		c.ServerError("GetRepoGitTree", err)
		return
	}

	templateURL := fmt.Sprintf("%s/repos/%s/%s/git/trees", c.BaseURL, c.Params(":username"), c.Params(":reponame"))

	if len(entries) == 0 {
		c.JSONSuccess(&repoGitTree{
			Sha: c.Params(":sha"),
			URL: fmt.Sprintf(templateURL+"/%s", c.Params(":sha")),
		})
		return
	}

	children := make([]*repoGitTreeEntry, 0, len(entries))
	for _, entry := range entries {
		var mode string
		switch entry.Type {
		case git.ObjectCommit:
			mode = "160000"
		case git.ObjectTree:
			mode = "040000"
		case git.ObjectBlob:
			mode = "120000"
		case git.ObjectTag:
			mode = "100644"
		default:
			mode = ""
		}
		children = append(children, &repoGitTreeEntry{
			Path: entry.Name(),
			Mode: mode,
			Type: string(entry.Type),
			Size: entry.Size(),
			Sha:  entry.ID.String(),
			URL:  fmt.Sprintf(templateURL+"/%s", entry.ID.String()),
		})
	}
	c.JSONSuccess(&repoGitTree{
		Sha:  c.Params(":sha"),
		URL:  fmt.Sprintf(templateURL+"/%s", c.Params(":sha")),
		Tree: children,
	})
}
