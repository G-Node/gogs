package user

import (
	"strings"

	"github.com/G-Node/gogs/models"
)

// excludeFromFeed returns 'true' if the given action should be excluded from the user feed.
func excludeFromFeed(act *models.Action) bool {
	return strings.Contains(act.RefName, "synced/git-annex") ||
		strings.Contains(act.RefName, "synced/master") ||
		strings.Contains(act.RefName, "git-annex")
}
