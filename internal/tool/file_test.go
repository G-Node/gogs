package tool

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_IsValidAnnexPointerFile(t *testing.T) {
	Convey("Check if a (file) content is a valid annex pointer file", t, func() {
		testCases := []struct {
			expect  bool
			content string
		}{
			// valid key and EOF
			{true, "/annex/objects/MD0-s232439--f4d0aaf2b2ac7a4cf00fbae9158a1b7c"},
			// not a key pattern
			{false, "foo/bar"},

			// key pattern doesn't start at the beginning of content
			{false, " /annex/objects/MD1-s232439--f4d0aaf2b2ac7a4cf00fbae9158a1b7c"},
			// key contain invalid character
			{false, "/annex/objects/M+D2-s232439--f4d0aaf2b2ac7a4cf00fbae9158a1b7c"},
			// newline after key (and no more content)
			{true, "/annex/objects/MD3-f4d0aaf2b2ac-7a4cf00fbae9158a1b7c\n"},
			// key can contains underscore (depending on backend)
			{true, "/annex/objects/SHA4_384-232439cf00fbae9158a1b7c"},

			// empty additional line
			{false, "/annex/objects/MD5-s232439--f4d0aaf2b2ac7a4cf00fbae9158a1b7c\n\n"},

			// valid additional line
			{true, "/annex/objects/MD6-f4d0aaf2ba4cf00fbae9158a1b7c\n/annex/\n"},
			// empty additional line
			{false, "/annex/objects/MD7-s232439--f4d0aaf2b2ac7a4cf00fbae9158a1b7c\n/annex/\n\n"},
			// additional line not terminated by new line
			{false, "/annex/objects/MD8-s232439--f4d0aaf2b2ac7a4cf00fbae9158a1b7c\n/annex/"},

			// valid additional lines
			{true, "/annex/objects/MD9-s232439--f4d0aaf2b2ac7a4cf00fbae9158a1b7c\r /annex/\n /annex/\n/annex/ \n"},
			// many valid additional lines, within the 32kb max file size
			{true, "/annex/objects/MD10-s232439--f4d0aaf2b2ac7a4cf00fbae9158a1b7c\n" + strings.Repeat("/annex/89\n", 31*1024/10)},
			// many valid additional lines, over the 32kb max file size
			{false, "/annex/objects/MD11-s232439--f4d0aaf2b2ac7a4cf00fbae9158a1b7c\n" + strings.Repeat("/annex/89\n", 32*1024/10)},
		}

		for _, tc := range testCases {
			So(IsAnnexedFile([]byte(tc.content)), ShouldEqual, tc.expect)
		}
	})
}
