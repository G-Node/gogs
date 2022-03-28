package user

import (
	"strings"

	"github.com/NII-DG/gogs/internal/db"
)

// excludeFromFeed returns 'true' if the given action should be excluded from the user feed.
func excludeFromFeed(act *db.Action) bool {
	return strings.Contains(act.RefName, "synced/git-annex") ||
		strings.Contains(act.RefName, "synced/master") ||
		strings.Contains(act.RefName, "git-annex")
}
