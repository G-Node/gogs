package route

import "github.com/G-Node/gogs/internal/db"

func filterUnlistedRepos(repos []*db.Repository) []*db.Repository {
	// Filter out Unlisted repositories
	var showRep []*db.Repository
	for _, repo := range repos {
		if !repo.IsUnlisted {
			showRep = append(showRep, repo)
		}
	}
	return showRep
}
