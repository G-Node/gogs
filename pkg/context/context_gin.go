package context

import (
	"os"
	"path"
	"strings"

	"github.com/G-Node/gogs/pkg/setting"
	"github.com/G-Node/gogs/pkg/tool"
	"github.com/Unknwon/com"
	log "gopkg.in/clog.v1"
)

// readNotice checks if a notice file exists and loads the message to display
// on all pages.
func readNotice(c *Context) {

	fileloc := path.Join(setting.CustomPath, "notice")
	var maxlen int64 = 1024

	if !com.IsExist(fileloc) {
		return
	}

	log.Trace("Found notice file")
	fp, err := os.Open(fileloc)
	if err != nil {
		log.Error(2, "Failed to open notice file %s: %v", fileloc, err)
		return
	}
	defer fp.Close()

	finfo, err := fp.Stat()
	if err != nil {
		log.Error(2, "Failed to stat notice file %s: %v", fileloc, err)
		return
	}

	if finfo.Size() > maxlen { // Refuse to print very long messages
		log.Error(2, "Notice file %s size too large [%d > %d]: refusing to render", fileloc, finfo.Size(), maxlen)
		return
	}

	buf := make([]byte, maxlen)
	n, err := fp.Read(buf)
	if err != nil {
		log.Error(2, "Failed to read notice file: %v", err)
		return
	}
	buf = buf[:n]

	if !tool.IsTextFile(buf) {
		log.Error(2, "Notice file %s does not appear to be a text file: aborting", fileloc)
		return
	}

	noticetext := strings.SplitN(string(buf), "\n", 2)
	c.Data["HasNotice"] = true
	c.Data["NoticeTitle"] = noticetext[0]
	c.Data["NoticeMessage"] = noticetext[1]
}
