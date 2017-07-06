package models_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	. "github.com/G-Node/gogs/models"
	"github.com/G-Node/gogs/pkg/markup"
)

func TestRepo(t *testing.T) {
	Convey("The metas map", t, func() {
		var repo = new(Repository)
		repo.Name = "testrepo"
		repo.Owner = new(User)
		repo.Owner.Name = "testuser"
		repo.ExternalTrackerFormat = "https://someurl.com/{user}/{repo}/{issue}"

		Convey("When no external tracker is configured", func() {
			Convey("It should be nil", func() {
				repo.EnableExternalTracker = false
				So(repo.ComposeMetas(), ShouldEqual, map[string]string(nil))
			})
			Convey("It should be nil even if other settings are present", func() {
				repo.EnableExternalTracker = false
				repo.ExternalTrackerFormat = "http://someurl.com/{user}/{repo}/{issue}"
				repo.ExternalTrackerStyle = markup.ISSUE_NAME_STYLE_NUMERIC
				So(repo.ComposeMetas(), ShouldEqual, map[string]string(nil))
			})
		})

		Convey("When an external issue tracker is configured", func() {
			repo.EnableExternalTracker = true
			Convey("It should default to numeric issue style", func() {
				metas := repo.ComposeMetas()
				So(metas["style"], ShouldEqual, markup.ISSUE_NAME_STYLE_NUMERIC)
			})
			Convey("It should pass through numeric issue style setting", func() {
				repo.ExternalTrackerStyle = markup.ISSUE_NAME_STYLE_NUMERIC
				metas := repo.ComposeMetas()
				So(metas["style"], ShouldEqual, markup.ISSUE_NAME_STYLE_NUMERIC)
			})
			Convey("It should pass through alphanumeric issue style setting", func() {
				repo.ExternalTrackerStyle = markup.ISSUE_NAME_STYLE_ALPHANUMERIC
				metas := repo.ComposeMetas()
				So(metas["style"], ShouldEqual, markup.ISSUE_NAME_STYLE_ALPHANUMERIC)
			})
			Convey("It should contain the user name", func() {
				metas := repo.ComposeMetas()
				So(metas["user"], ShouldEqual, "testuser")
			})
			Convey("It should contain the repo name", func() {
				metas := repo.ComposeMetas()
				So(metas["repo"], ShouldEqual, "testrepo")
			})
			Convey("It should contain the URL format", func() {
				metas := repo.ComposeMetas()
				So(metas["format"], ShouldEqual, "https://someurl.com/{user}/{repo}/{issue}")
			})
		})
	})
}
