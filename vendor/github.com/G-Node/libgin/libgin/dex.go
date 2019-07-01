package libgin

import (
	"time"

	"github.com/G-Node/gig"
)

// NOTE: TEMPORARY COPY FROM gin-dex

type SearchRequest struct {
	Keywords string
	SType    int
	RepoIDs  []int64
}

const (
	SEARCH_MATCH = iota
	SEARCH_FUZZY
	SEARCH_WILDCARD
	SEARCH_QUERY
	SEARCH_SUGGEST
)

type BlobSResult struct {
	Source    *IndexBlob  `json:"_source"`
	Score     float64     `json:"_score"`
	Highlight interface{} `json:"highlight"`
}

type CommitSResult struct {
	Source    *IndexCommit `json:"_source"`
	Score     float64      `json:"_score"`
	Highlight interface{}  `json:"highlight"`
}

type SearchResults struct {
	Blobs   []BlobSResult
	Commits []CommitSResult
}

type IndexBlob struct {
	*gig.Blob
	GinRepoName  string
	GinRepoId    string
	FirstCommit  string
	Id           int64
	Oid          gig.SHA1
	IndexingTime time.Time
	Content      string
	Path         string
}

type IndexCommit struct {
	*gig.Commit
	GinRepoId    string
	Oid          gig.SHA1
	GinRepoName  string
	IndexingTime time.Time
}

type IndexRequest struct {
	UserID   int64
	RepoID   int64
	RepoPath string
}
