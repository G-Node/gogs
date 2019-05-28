package libgin

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// NOTE: TEMPORARY COPIES FROM gin-doi

// uuidMap is a map between registered repositories and their UUIDs for datasets registered before the new UUID generation method was implemented.
// This map is required because the current method of computing UUIDs differs from the older method and this lookup is used to handle the old-method UUIDs.
var uuidMap = map[string]string{
	"INT/multielectrode_grasp":                   "f83565d148510fede8a277f660e1a419",
	"ajkumaraswamy/HB-PAC_disinhibitory_network": "1090f803258557299d287c4d44a541b2",
	"steffi/Kleineidam_et_al_2017":               "f53069de4c4921a3cfa8f17d55ef98bb",
	"Churan/Morris_et_al_Frontiers_2016":         "97bc1456d3f4bca2d945357b3ec92029",
	"fabee/efish_locking":                        "6953bbf0087ba444b2d549b759de4a06",
}

// RepoPathToUUID computes a UUID from a repository path.
func RepoPathToUUID(URI string) string {
	if doi, ok := uuidMap[URI]; ok {
		return doi
	}
	currMd5 := md5.Sum([]byte(URI))
	return hex.EncodeToString(currMd5[:])
}

// DOIRegInfo holds all the metadata and information necessary for a DOI registration request.
type DOIRegInfo struct {
	Missing      []string
	DOI          string
	UUID         string
	FileSize     int64
	Title        string
	Authors      []Author
	Description  string
	Keywords     []string
	References   []Reference
	Funding      []string
	License      *License
	ResourceType string
}

func (c *DOIRegInfo) GetType() string {
	if c.ResourceType != "" {
		return c.ResourceType
	}
	return "Dataset"
}

func (c *DOIRegInfo) GetCitation() string {
	var authors string
	for _, auth := range c.Authors {
		if len(auth.FirstName) > 0 {
			authors += fmt.Sprintf("%s %s, ", auth.LastName, string(auth.FirstName[0]))
		} else {
			authors += fmt.Sprintf("%s, ", auth.LastName)
		}
	}
	return fmt.Sprintf("%s (%d) %s. G-Node. doi:%s", authors, time.Now().Year(), c.Title, c.DOI)
}

func (c *DOIRegInfo) EscXML(txt string) string {
	buf := new(bytes.Buffer)
	if err := xml.EscapeText(buf, []byte(txt)); err != nil {
		log.Printf("Could not escape: %s :: %s", txt, err.Error())
		return ""
	}
	return buf.String()

}

type Author struct {
	FirstName   string
	LastName    string
	Affiliation string
	ID          string
}

func (c *Author) GetValidID() *NamedIdentifier {
	if c.ID == "" {
		return nil
	}
	if strings.Contains(strings.ToLower(c.ID), "orcid") {
		// assume the orcid id is a four block number thing eg. 0000-0002-5947-9939
		var re = regexp.MustCompile(`(\d+-\d+-\d+-\d+)`)
		nid := string(re.Find([]byte(c.ID)))
		return &NamedIdentifier{URI: "https://orcid.org/", Scheme: "ORCID", ID: nid}
	}
	return nil
}
func (a *Author) RenderAuthor() string {
	auth := fmt.Sprintf("%s,%s;%s;%s", a.LastName, a.FirstName, a.Affiliation, a.ID)
	return strings.TrimRight(auth, ";")
}

type NamedIdentifier struct {
	URI    string
	Scheme string
	ID     string
}

type Reference struct {
	Reftype string
	Name    string
	ID      string
}

type License struct {
	Name string
	URL  string
}

func IsRegisteredDOI(doi string) bool {
	url := fmt.Sprintf("https://doi.org/%s", doi)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Could not query for DOI: %s at %s", doi, url)
		return false
	}
	if resp.StatusCode != http.StatusNotFound {
		return true
	}
	return false
}
