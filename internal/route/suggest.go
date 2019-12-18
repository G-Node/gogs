package route

import (
	"github.com/G-Node/gogs/internal/context"
	"github.com/G-Node/libgin/libgin"
	log "gopkg.in/clog.v1"
)

// ExploreSuggest returns suggestions for keywords to fill the search box on the explore/data page.
func ExploreSuggest(c *context.Context) {
	keywords := c.Params(":keywords")
	sType := libgin.SEARCH_SUGGEST
	log.Trace("Suggestions for [%s]", keywords)

	if keywords == "" {
		// no keywords submitted: return
		return
	}
	// res := libgin.SearchResults{}

	log.Trace("Searching data/blobs for suggestions")
	data, err := search(c, keywords, sType)
	if err != nil {
		log.Error(2, "Query returned error: %v", err)
		return
	}

	log.Trace("Returning suggestions: %+v", string(data))

	if err != nil {
		log.Error(2, "Failed to marshal structured suggestions: %v", err)
		return
	}

	c.Write(data)
}
