package gindex

import (
	"github.com/G-Node/gig"
	log "github.com/Sirupsen/logrus"
)

func IndexRepoWithPath(path, ref string, serv *ElServer, repoid string, reponame string) error {
	log.Info("Start indexing repository with path: %s", path)
	rep, err := gig.OpenRepository(path)
	if err != nil {
		log.Errorf("Could not open repository: %+v", err)
		return err
	}
	log.Debugf("Opened repository")
	commits, err := rep.WalkRef(ref,
		func(comitID gig.SHA1) bool {
			res, err := serv.HasCommit("commits", GetIndexCommitId(comitID.String(), repoid))
			if err != nil {
				log.Errorf("Could not querry commit index: %v", err)
				return false
			}
			return !res
		})
	log.Infof("Found %d commits", len(commits))

	for commitid, commit := range commits {
		err = indexCommit(commit, repoid, commitid, rep, path, reponame, serv, serv.HasBlob)
	}
	return nil
}

func ReIndexRepoWithPath(path, ref string, serv *ElServer, repoid string, reponame string) error {
	log.Info("Start indexing repository with path: %s", path)
	rep, err := gig.OpenRepository(path)
	if err != nil {
		log.Errorf("Could not open repository: %+v", err)
		return err
	}
	log.Debugf("Opened repository")
	commits, err := rep.WalkRef(ref,
		func(comitID gig.SHA1) bool {
			return true
		})
	log.Infof("Found %d commits", len(commits))

	blobs := make(map[gig.SHA1]bool)
	for commitid, commit := range commits {
		err = indexCommit(commit, repoid, commitid, rep, path, reponame, serv,
			func(indexName string, id gig.SHA1) (bool, error) {
				if !blobs[id] {
					blobs[id] = true
					return false, nil
				}
				return true, nil
			})
	}
	return nil
}

func indexCommit(commit *gig.Commit, repoid string, commitid gig.SHA1, rep *gig.Repository,
	path string, reponame string, serv *ElServer,
	indexBlob func(string, gig.SHA1) (bool, error)) error {
	err := NewCommitFromGig(commit, repoid, reponame, commitid).AddToIndex(serv, "commits", commitid)
	if err != nil {
		log.Printf("Indexing commit failed:%+v", err)
	}
	blobs := make(map[gig.SHA1]*gig.Blob)
	rep.GetBlobsForCommit(commit, blobs)
	for blid, blob := range blobs {
		log.Debugf("Blob %s has Size:%d", blid, blob.Size())
		hasBlob, err := indexBlob("blobs", GetIndexBlobId(blid.String(), repoid))
		if err != nil {
			log.Errorf("Could not querry for blob: %+v", err)
			return err
		}
		if !hasBlob {
			bpath, _ := GetBlobPath(blid.String(), commitid.String(), path)
			err = NewBlobFromGig(blob, repoid, blid, commitid.String(), bpath, reponame).AddToIndex(serv, "blobs", path, blid)
			if err != nil {
				log.Debugf("Indexing blob failed: %+v", err)
			}
		} else {
			log.Debugf("Blob there :%s", blid)
		}
	}
	return nil
}