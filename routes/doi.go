package routes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"

	"github.com/G-Node/gogs/pkg/context"
	"github.com/G-Node/gogs/pkg/setting"
	log "gopkg.in/clog.v1"
)

func RequestDOI(c *context.Context) {
	if !c.Repo.IsAdmin() {
		c.Status(http.StatusUnauthorized)
		return
	}
	token := c.GetCookie(setting.SessionConfig.CookieName)
	token, err := encrypt([]byte(setting.DOI.Key), token)
	if err != nil {
		log.Error(2, "Could not encrypt token for DOI request: %s", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	doiurl, err := url.Parse(setting.DOI.URL + "/register") // TODO: Handle error by notifying admin email
	if err != nil {
		log.Error(2, "Failed to parse DOI URL: %s", setting.DOI.URL)
	}

	params := url.Values{}
	params.Add("repo", c.Repo.Repository.FullName())
	params.Add("user", c.User.Name)
	params.Add("token", token)
	doiurl.RawQuery = params.Encode()
	target, _ := url.PathUnescape(doiurl.String())
	log.Trace(target)
	c.Redirect(target)
}

// NOTE: TEMPORARY COPY FROM gin-doi

// encrypt string to base64 crypto using AES
func encrypt(key []byte, text string) (string, error) {
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}
