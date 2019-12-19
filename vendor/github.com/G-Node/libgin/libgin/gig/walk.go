package gig

import "fmt"

func (repo *Repository) WalkRef(refname string, goOn func(SHA1) bool) (map[SHA1]*Commit, error) {
	head, err := repo.OpenRef(refname)
	if err != nil {
		return nil, err
	}

	HId, err := head.Resolve()
	if err != nil {
		return nil, err
	}

	commits := make(map[SHA1]*Commit)
	repo.walkCommitTree(commits, HId, goOn)
	return commits, nil
}

func (repo *Repository) walkCommitTree(commits map[SHA1]*Commit, commitId SHA1,
	goOn func(SHA1) bool) error {
	commit, err := repo.OpenObject(commitId)
	commit.Close()
	if err != nil {
		return err
	}

	if _, ok := commits[commitId]; !ok && goOn(commitId) {
		commits[commitId] = commit.(*Commit)
		for _, parent := range commit.(*Commit).Parent {
			repo.walkCommitTree(commits, parent, goOn)
		}
		return nil
	} else {
		return nil
	}
}

func (repo *Repository) GetBlobsForCommit(commit *Commit, blobs map[SHA1]*Blob) error {
	treeOb, err := repo.OpenObject(commit.Tree)
	if err != nil {
		return err
	}
	defer treeOb.Close()

	tree, ok := treeOb.(*Tree)
	if !ok {
		return fmt.Errorf("Could not assert a tree")
	}

	err = repo.GetBlobsForTree(tree, blobs)
	return err
}

func (repo *Repository) GetBlobsForTree(tree *Tree, blobs map[SHA1]*Blob) error {
	for tree.Next() {
		trEntry := tree.Entry()
		switch trEntry.Type {
		case ObjBlob:
			if blobOb, err := repo.OpenObject(trEntry.ID); err != nil {
				return err
			} else {
				blobs[trEntry.ID] = blobOb.(*Blob)
				blobOb.Close()
			}
		case ObjTree:
			if treeOb, err := repo.OpenObject(trEntry.ID); err != nil {
				return err
			} else {
				if err = repo.GetBlobsForTree(treeOb.(*Tree), blobs); err != nil {
					treeOb.Close()
					return err
				}

			}
		}
	}
	return tree.Err()
}
