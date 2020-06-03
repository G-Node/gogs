package context

import (
	"strings"

	"github.com/G-Node/git-module"
	"github.com/G-Node/gogs/internal/db"
	"github.com/G-Node/gogs/internal/setting"
	"github.com/G-Node/libgin/libgin"
	log "gopkg.in/clog.v1"
)

// getRepoDOI returns the DOI for the repository based on the following rules:
// - if the repository belongs to the DOI user and has a tag that matches the
// DOI prefix, returns the tag.
// - if the repo is forked by the DOI user, check the DOI fork for the tag as above.
// - if the repo is forked by the DOI user and the fork doesn't have a tag,
// returns the (old-style) calculated DOI, based on the hash of the repository
// path.
// - An empty string is returned if it is not not forked by the DOI user.
// If an error occurs at any point, returns an empty string (the error is logged).
// Tag retrieval is allowed to fail and falls back on the hashed DOI method.
func getRepoDOI(c *Context) string {
	repo := c.Repo.Repository
	var doiFork *db.Repository
	if repo.Owner.Name == "doi" {
		doiFork = repo
	} else {
		if forks, err := repo.GetForks(); err == nil {
			for _, fork := range forks {
				if fork.MustOwner().Name == "doi" {
					doiFork = fork
					break
				}
			}
		} else {
			log.Error(2, "failed to get forks for repository %q (%d): %v", repo.FullName(), repo.ID, err)
			return ""
		}
	}

	if doiFork == nil {
		// not owned or forked by DOI, so not registered
		return ""
	}

	// check the DOI fork for a tag that matches our DOI prefix
	// if multiple exit, get the latest one
	doiBase := setting.DOI.Base

	doiForkGit, err := git.OpenRepository(doiFork.RepoPath())
	if err != nil {
		log.Error(2, "failed to open git repository at %q (%d): %v", doiFork.RepoPath(), doiFork.ID, err)
		return ""
	}
	if tags, err := doiForkGit.GetTags(); err == nil {
		var latestTime int64
		latestTag := ""
		for _, tagName := range tags {
			if strings.Contains(tagName, doiBase) {
				tag, err := doiForkGit.GetTag(tagName)
				if err != nil {
					// log the error and continue to the next tag
					log.Error(2, "failed to get information for tag %q for repository at %q: %v", tagName, doiForkGit.Path, err)
					continue
				}
				commit, err := tag.Commit()
				if err != nil {
					// log the error and continue to the next tag
					log.Error(2, "failed to get commit for tag %q for repository at %q: %v", tagName, doiForkGit.Path, err)
					continue
				}
				commitTime := commit.Committer.When.Unix()
				if commitTime > latestTime {
					latestTag = tagName
					latestTime = commitTime
				}
				return latestTag
			}
		}
	} else {
		// this shouldn't happen even if there are no tags
		// log the error, but fall back to the old method anyway
		log.Error(2, "failed to get tags for repository at %q: %v", doiForkGit.Path, err)
	}

	// Has DOI fork but isn't tagged: return old style has-based DOI
	repoPath := repo.FullName()
	// get base repo name if it's a DOI fork
	if c.Repo.Repository.IsFork && c.Repo.Owner.Name == "doi" {
		repoPath = c.Repo.Repository.BaseRepo.FullName()
	}
	uuid := libgin.RepoPathToUUID(repoPath)
	return doiBase + uuid[:6]
}
