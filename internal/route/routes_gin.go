package routes

import "github.com/G-Node/gogs/models"

func filterUnlistedRepos(repos []*models.Repository) []*models.Repository {
	// Filter out Unlisted repositories
	var showRep []*models.Repository
	for _, repo := range repos {
		if !repo.Unlisted {
			showRep = append(showRep, repo)
		}
	}
	return showRep
}
