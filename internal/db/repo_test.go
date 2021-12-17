package db

import (
	_ "image/jpeg"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ivis-yoshida/gogs/internal/markup"
	"github.com/stretchr/testify/assert"
	"github.com/unknwon/com"
)

func TestRepository_ComposeMetas(t *testing.T) {
	repo := &Repository{
		Name: "testrepo",
		Owner: &User{
			Name: "testuser",
		},
		ExternalTrackerFormat: "https://someurl.com/{user}/{repo}/{issue}",
	}

	t.Run("no external tracker is configured", func(t *testing.T) {
		repo.EnableExternalTracker = false

		metas := repo.ComposeMetas()
		assert.Equal(t, metas["repoLink"], repo.Link())

		// Should no format and style if no external tracker is configured
		_, ok := metas["format"]
		assert.False(t, ok)
		_, ok = metas["style"]
		assert.False(t, ok)
	})

	t.Run("an external issue tracker is configured", func(t *testing.T) {
		repo.ExternalMetas = nil
		repo.EnableExternalTracker = true

		// Default to numeric issue style
		assert.Equal(t, markup.ISSUE_NAME_STYLE_NUMERIC, repo.ComposeMetas()["style"])
		repo.ExternalMetas = nil

		repo.ExternalTrackerStyle = markup.ISSUE_NAME_STYLE_NUMERIC
		assert.Equal(t, markup.ISSUE_NAME_STYLE_NUMERIC, repo.ComposeMetas()["style"])
		repo.ExternalMetas = nil

		repo.ExternalTrackerStyle = markup.ISSUE_NAME_STYLE_ALPHANUMERIC
		assert.Equal(t, markup.ISSUE_NAME_STYLE_ALPHANUMERIC, repo.ComposeMetas()["style"])
		repo.ExternalMetas = nil

		metas := repo.ComposeMetas()
		assert.Equal(t, "testuser", metas["user"])
		assert.Equal(t, "testrepo", metas["repo"])
		assert.Equal(t, "https://someurl.com/{user}/{repo}/{issue}", metas["format"])
	})
}

func Test_initWorkflow(t *testing.T) {
	repo := &Repository{
		Name: "testrepo",
		Owner: &User{
			Name: "testuser",
		},
		ExternalTrackerFormat: "https://someurl.com/{user}/{repo}/{issue}",
	}

	type args struct {
		tmpDir, workflowUrl string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "noError",
			args: args{
				tmpDir:      filepath.Join(os.TempDir(), "gogs-"+repo.Name+"-"+com.ToStr(time.Now().Nanosecond())),
				workflowUrl: "https://github.com/ivis-kuwata/workflow-template",
			},
			wantErr: false,
		},
		{
			name: "fail git clone",
			args: args{
				tmpDir:      filepath.Join(os.TempDir(), "gogs-"+repo.Name+"-"+com.ToStr(time.Now().Nanosecond())),
				workflowUrl: "https://github.com/noexist/workflow-template",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := initWorkflow(tt.args.tmpDir, tt.args.workflowUrl); (err != nil) != tt.wantErr {
				t.Errorf("initWorkflow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
