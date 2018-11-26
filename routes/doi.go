package routes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/G-Node/gogs/pkg/context"
	"github.com/G-Node/gogs/pkg/setting"
	log "gopkg.in/clog.v1"
)

func RequestDoi(c *context.Context) {
	if !c.Repo.IsAdmin() {
		c.Status(http.StatusUnauthorized)
		return
	}
	token := c.GetCookie(setting.SessionConfig.CookieName)
	token, err := encrypt([]byte(setting.Doi.DoiKey), token)
	if err != nil {
		log.Error(0, "Could not encrypt Secret key:%s", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	url := fmt.Sprintf("%s/?repo=%s&user=%s&token=%s", setting.Doi.DoiUrl, c.Repo.Repository.FullName(),
		c.User.Name, token)
	c.Redirect(url)
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
