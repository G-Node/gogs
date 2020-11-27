// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"fmt"

	log "unknwon.dev/clog/v2"
)

type AccessMode int

const (
	AccessModeNone  AccessMode = iota // 0
	AccessModeRead                    // 1
	AccessModeWrite                   // 2
	AccessModeAdmin                   // 3
	AccessModeOwner                   // 4
)

func (mode AccessMode) String() string {
	switch mode {
	case AccessModeRead:
		return "read"
	case AccessModeWrite:
		return "write"
	case AccessModeAdmin:
		return "admin"
	case AccessModeOwner:
		return "owner"
	default:
		return "none"
	}
}

// ParseAccessMode returns corresponding access mode to given permission string.
func ParseAccessMode(permission string) AccessMode {
	switch permission {
	case "write":
		return AccessModeWrite
	case "admin":
		return AccessModeAdmin
	default:
		return AccessModeRead
	}
}

// Access represents the highest access level of a user to a repository. The only access type
// that is not in this table is the real owner of a repository. In case of an organization
// repository, the members of the owners team are in this table.
type Access struct {
	ID     int64
	UserID int64 `xorm:"UNIQUE(s)"`
	RepoID int64 `xorm:"UNIQUE(s)"`
	Mode   AccessMode
}

func userAccessMode(e Engine, userID int64, repo *Repository) (AccessMode, error) {
	mode := AccessModeNone
	// Everyone has read access to public repository
	if !repo.IsPrivate {
		mode = AccessModeRead
	}

	if userID <= 0 {
		return mode, nil
	}

	if userID == repo.OwnerID {
		return AccessModeOwner, nil
	}

	access := &Access{
		UserID: userID,
		RepoID: repo.ID,
	}
	if has, err := e.Get(access); !has || err != nil {
		return mode, err
	}
	return access.Mode, nil
}

// UserAccessMode returns the access mode of given user to the repository.
func UserAccessMode(userID int64, repo *Repository) (AccessMode, error) {
	return userAccessMode(x, userID, repo)
}

func hasAccess(e Engine, userID int64, repo *Repository, testMode AccessMode) (bool, error) {
	mode, err := userAccessMode(e, userID, repo)
	return mode >= testMode, err
}

// HasAccess returns true if someone has the request access level. User can be nil!
// Deprecated: Use Perms.Authorize instead.
func HasAccess(userID int64, repo *Repository, testMode AccessMode) (bool, error) {
	return hasAccess(x, userID, repo, testMode)
}

// GetRepositoryAccesses finds all repositories with their access mode where a user has access but does not own.
func (u *User) GetRepositoryAccesses() (map[*Repository]AccessMode, error) {
	accesses := make([]*Access, 0, 10)
	if err := x.Find(&accesses, &Access{UserID: u.ID}); err != nil {
		return nil, err
	}

	repos := make(map[*Repository]AccessMode, len(accesses))
	for _, access := range accesses {
		repo, err := GetRepositoryByID(access.RepoID)
		if err != nil {
			if IsErrRepoNotExist(err) {
				log.Error("Failed to get repository by ID: %v", err)
				continue
			}
			return nil, err
		}
		if repo.OwnerID == u.ID {
			continue
		}
		repos[repo] = access.Mode
	}
	return repos, nil
}

// GetAccessibleRepositories finds repositories which the user has access but does not own.
// If limit is smaller than 1 means returns all found results.
func (user *User) GetAccessibleRepositories(limit int) (repos []*Repository, _ error) {
	sess := x.Where("owner_id !=? ", user.ID).Desc("updated_unix")
	if limit > 0 {
		sess.Limit(limit)
		repos = make([]*Repository, 0, limit)
	} else {
		repos = make([]*Repository, 0, 10)
	}
	return repos, sess.Join("INNER", "access", "access.user_id = ? AND access.repo_id = repository.id", user.ID).Find(&repos)
}

func maxAccessMode(modes ...AccessMode) AccessMode {
	max := AccessModeNone
	for _, mode := range modes {
		if mode > max {
			max = mode
		}
	}
	return max
}

// Deprecated: Use Perms.SetRepoPerms instead.
func (repo *Repository) refreshAccesses(e Engine, accessMap map[int64]AccessMode) (err error) {
	newAccesses := make([]Access, 0, len(accessMap))
	for userID, mode := range accessMap {
		newAccesses = append(newAccesses, Access{
			UserID: userID,
			RepoID: repo.ID,
			Mode:   mode,
		})
	}

	// Delete old accesses and insert new ones for repository.
	if _, err = e.Delete(&Access{RepoID: repo.ID}); err != nil {
		return fmt.Errorf("delete old accesses: %v", err)
	} else if _, err = e.Insert(newAccesses); err != nil {
		return fmt.Errorf("insert new accesses: %v", err)
	}
	return nil
}

// refreshCollaboratorAccesses retrieves repository collaborations with their access modes.
func (repo *Repository) refreshCollaboratorAccesses(e Engine, accessMap map[int64]AccessMode) error {
	collaborations, err := repo.getCollaborations(e)
	if err != nil {
		return fmt.Errorf("getCollaborations: %v", err)
	}
	for _, c := range collaborations {
		accessMap[c.UserID] = c.Mode
	}
	return nil
}

// recalculateTeamAccesses recalculates new accesses for teams of an organization
// except the team whose ID is given. It is used to assign a team ID when
// remove repository from that team.
func (repo *Repository) recalculateTeamAccesses(e Engine, ignTeamID int64) (err error) {
	accessMap := make(map[int64]AccessMode, 20)

	if err = repo.getOwner(e); err != nil {
		return err
	} else if !repo.Owner.IsOrganization() {
		return fmt.Errorf("owner is not an organization: %d", repo.OwnerID)
	}

	if err = repo.refreshCollaboratorAccesses(e, accessMap); err != nil {
		return fmt.Errorf("refreshCollaboratorAccesses: %v", err)
	}

	if err = repo.Owner.getTeams(e); err != nil {
		return err
	}

	for _, t := range repo.Owner.Teams {
		if t.ID == ignTeamID {
			continue
		}

		// Owner team gets owner access, and skip for teams that do not
		// have relations with repository.
		if t.IsOwnerTeam() {
			t.Authorize = AccessModeOwner
		} else if !t.hasRepository(e, repo.ID) {
			continue
		}

		if err = t.getMembers(e); err != nil {
			return fmt.Errorf("getMembers '%d': %v", t.ID, err)
		}
		for _, m := range t.Members {
			accessMap[m.ID] = maxAccessMode(accessMap[m.ID], t.Authorize)
		}
	}

	return repo.refreshAccesses(e, accessMap)
}

func (repo *Repository) recalculateAccesses(e Engine) error {
	if repo.Owner.IsOrganization() {
		return repo.recalculateTeamAccesses(e, 0)
	}

	accessMap := make(map[int64]AccessMode, 10)
	if err := repo.refreshCollaboratorAccesses(e, accessMap); err != nil {
		return fmt.Errorf("refreshCollaboratorAccesses: %v", err)
	}
	return repo.refreshAccesses(e, accessMap)
}

// RecalculateAccesses recalculates all accesses for repository.
func (repo *Repository) RecalculateAccesses() error {
	return repo.recalculateAccesses(x)
}
