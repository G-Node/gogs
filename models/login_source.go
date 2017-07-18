// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/Unknwon/com"
	"github.com/go-macaron/binding"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	log "gopkg.in/clog.v1"

	"github.com/G-Node/gogs/models/errors"
	"github.com/G-Node/gogs/pkg/auth/ldap"
	"github.com/G-Node/gogs/pkg/auth/pam"
)

type LoginType int

// Note: new type must append to the end of list to maintain compatibility.
const (
	LOGIN_NOTYPE LoginType = iota
	LOGIN_PLAIN            // 1
	LOGIN_LDAP             // 2
	LOGIN_SMTP             // 3
	LOGIN_PAM              // 4
	LOGIN_DLDAP            // 5
)

var LoginNames = map[LoginType]string{
	LOGIN_LDAP:  "LDAP (via BindDN)",
	LOGIN_DLDAP: "LDAP (simple auth)", // Via direct bind
	LOGIN_SMTP:  "SMTP",
	LOGIN_PAM:   "PAM",
}

var SecurityProtocolNames = map[ldap.SecurityProtocol]string{
	ldap.SECURITY_PROTOCOL_UNENCRYPTED: "Unencrypted",
	ldap.SECURITY_PROTOCOL_LDAPS:       "LDAPS",
	ldap.SECURITY_PROTOCOL_START_TLS:   "StartTLS",
}

// Ensure structs implemented interface.
var (
	_ core.Conversion = &LDAPConfig{}
	_ core.Conversion = &SMTPConfig{}
	_ core.Conversion = &PAMConfig{}
)

type LDAPConfig struct {
	*ldap.Source
}

func (cfg *LDAPConfig) FromDB(bs []byte) error {
	return json.Unmarshal(bs, &cfg)
}

func (cfg *LDAPConfig) ToDB() ([]byte, error) {
	return json.Marshal(cfg)
}

func (cfg *LDAPConfig) SecurityProtocolName() string {
	return SecurityProtocolNames[cfg.SecurityProtocol]
}

type SMTPConfig struct {
	Auth           string
	Host           string
	Port           int
	AllowedDomains string `xorm:"TEXT"`
	TLS            bool
	SkipVerify     bool
}

func (cfg *SMTPConfig) FromDB(bs []byte) error {
	return json.Unmarshal(bs, cfg)
}

func (cfg *SMTPConfig) ToDB() ([]byte, error) {
	return json.Marshal(cfg)
}

type PAMConfig struct {
	ServiceName string // pam service (e.g. system-auth)
}

func (cfg *PAMConfig) FromDB(bs []byte) error {
	return json.Unmarshal(bs, &cfg)
}

func (cfg *PAMConfig) ToDB() ([]byte, error) {
	return json.Marshal(cfg)
}

// LoginSource represents an external way for authorizing users.
type LoginSource struct {
	ID        int64
	Type      LoginType
	Name      string          `xorm:"UNIQUE"`
	IsActived bool            `xorm:"NOT NULL DEFAULT false"`
	Cfg       core.Conversion `xorm:"TEXT"`

	Created     time.Time `xorm:"-"`
	CreatedUnix int64
	Updated     time.Time `xorm:"-"`
	UpdatedUnix int64
}

func (s *LoginSource) BeforeInsert() {
	s.CreatedUnix = time.Now().Unix()
	s.UpdatedUnix = s.CreatedUnix
}

func (s *LoginSource) BeforeUpdate() {
	s.UpdatedUnix = time.Now().Unix()
}

// Cell2Int64 converts a xorm.Cell type to int64,
// and handles possible irregular cases.
func Cell2Int64(val xorm.Cell) int64 {
	switch (*val).(type) {
	case []uint8:
		log.Trace("Cell2Int64 ([]uint8): %v", *val)
		return com.StrTo(string((*val).([]uint8))).MustInt64()
	}
	return (*val).(int64)
}

func (source *LoginSource) BeforeSet(colName string, val xorm.Cell) {
	switch colName {
	case "type":
		switch LoginType(Cell2Int64(val)) {
		case LOGIN_LDAP, LOGIN_DLDAP:
			source.Cfg = new(LDAPConfig)
		case LOGIN_SMTP:
			source.Cfg = new(SMTPConfig)
		case LOGIN_PAM:
			source.Cfg = new(PAMConfig)
		default:
			panic("unrecognized login source type: " + com.ToStr(*val))
		}
	}
}

func (s *LoginSource) AfterSet(colName string, _ xorm.Cell) {
	switch colName {
	case "created_unix":
		s.Created = time.Unix(s.CreatedUnix, 0).Local()
	case "updated_unix":
		s.Updated = time.Unix(s.UpdatedUnix, 0).Local()
	}
}

func (source *LoginSource) TypeName() string {
	return LoginNames[source.Type]
}

func (source *LoginSource) IsLDAP() bool {
	return source.Type == LOGIN_LDAP
}

func (source *LoginSource) IsDLDAP() bool {
	return source.Type == LOGIN_DLDAP
}

func (source *LoginSource) IsSMTP() bool {
	return source.Type == LOGIN_SMTP
}

func (source *LoginSource) IsPAM() bool {
	return source.Type == LOGIN_PAM
}

func (source *LoginSource) HasTLS() bool {
	return ((source.IsLDAP() || source.IsDLDAP()) &&
		source.LDAP().SecurityProtocol > ldap.SECURITY_PROTOCOL_UNENCRYPTED) ||
		source.IsSMTP()
}

func (source *LoginSource) UseTLS() bool {
	switch source.Type {
	case LOGIN_LDAP, LOGIN_DLDAP:
		return source.LDAP().SecurityProtocol != ldap.SECURITY_PROTOCOL_UNENCRYPTED
	case LOGIN_SMTP:
		return source.SMTP().TLS
	}

	return false
}

func (source *LoginSource) SkipVerify() bool {
	switch source.Type {
	case LOGIN_LDAP, LOGIN_DLDAP:
		return source.LDAP().SkipVerify
	case LOGIN_SMTP:
		return source.SMTP().SkipVerify
	}

	return false
}

func (source *LoginSource) LDAP() *LDAPConfig {
	return source.Cfg.(*LDAPConfig)
}

func (source *LoginSource) SMTP() *SMTPConfig {
	return source.Cfg.(*SMTPConfig)
}

func (source *LoginSource) PAM() *PAMConfig {
	return source.Cfg.(*PAMConfig)
}
func CreateLoginSource(source *LoginSource) error {
	has, err := x.Get(&LoginSource{Name: source.Name})
	if err != nil {
		return err
	} else if has {
		return ErrLoginSourceAlreadyExist{source.Name}
	}

	_, err = x.Insert(source)
	return err
}

func LoginSources() ([]*LoginSource, error) {
	auths := make([]*LoginSource, 0, 5)
	return auths, x.Find(&auths)
}

// GetLoginSourceByID returns login source by given ID.
func GetLoginSourceByID(id int64) (*LoginSource, error) {
	source := new(LoginSource)
	has, err := x.Id(id).Get(source)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrLoginSourceNotExist{id}
	}
	return source, nil
}

func UpdateSource(source *LoginSource) error {
	_, err := x.Id(source.ID).AllCols().Update(source)
	return err
}

func DeleteSource(source *LoginSource) error {
	count, err := x.Count(&User{LoginSource: source.ID})
	if err != nil {
		return err
	} else if count > 0 {
		return ErrLoginSourceInUse{source.ID}
	}
	_, err = x.Id(source.ID).Delete(new(LoginSource))
	return err
}

// CountLoginSources returns number of login sources.
func CountLoginSources() int64 {
	count, _ := x.Count(new(LoginSource))
	return count
}

// .____     ________      _____ __________
// |    |    \______ \    /  _  \\______   \
// |    |     |    |  \  /  /_\  \|     ___/
// |    |___  |    `   \/    |    \    |
// |_______ \/_______  /\____|__  /____|
//         \/        \/         \/

func composeFullName(firstname, surname, username string) string {
	switch {
	case len(firstname) == 0 && len(surname) == 0:
		return username
	case len(firstname) == 0:
		return surname
	case len(surname) == 0:
		return firstname
	default:
		return firstname + " " + surname
	}
}

// LoginViaLDAP queries if login/password is valid against the LDAP directory pool,
// and create a local user if success when enabled.
func LoginViaLDAP(user *User, login, password string, source *LoginSource, autoRegister bool) (*User, error) {
	username, fn, sn, mail, isAdmin, succeed := source.Cfg.(*LDAPConfig).SearchEntry(login, password, source.Type == LOGIN_DLDAP)
	if !succeed {
		// User not in LDAP, do nothing
		return nil, errors.UserNotExist{0, login}
	}

	if !autoRegister {
		return user, nil
	}

	// Fallback.
	if len(username) == 0 {
		username = login
	}
	// Validate username make sure it satisfies requirement.
	if binding.AlphaDashDotPattern.MatchString(username) {
		return nil, fmt.Errorf("Invalid pattern for attribute 'username' [%s]: must be valid alpha or numeric or dash(-_) or dot characters", username)
	}

	if len(mail) == 0 {
		mail = fmt.Sprintf("%s@localhost", username)
	}

	user = &User{
		LowerName:   strings.ToLower(username),
		Name:        username,
		FullName:    composeFullName(fn, sn, username),
		Email:       mail,
		LoginType:   source.Type,
		LoginSource: source.ID,
		LoginName:   login,
		IsActive:    true,
		IsAdmin:     isAdmin,
	}

	ok, err := IsUserExist(0, user.Name)
	if err != nil {
		return user, err
	}

	if ok {
		return user, UpdateUser(user)
	}

	return user, CreateUser(user)
}

//   _________   __________________________
//  /   _____/  /     \__    ___/\______   \
//  \_____  \  /  \ /  \|    |    |     ___/
//  /        \/    Y    \    |    |    |
// /_______  /\____|__  /____|    |____|
//         \/         \/

type smtpLoginAuth struct {
	username, password string
}

func (auth *smtpLoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(auth.username), nil
}

func (auth *smtpLoginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(auth.username), nil
		case "Password:":
			return []byte(auth.password), nil
		}
	}
	return nil, nil
}

const (
	SMTP_PLAIN = "PLAIN"
	SMTP_LOGIN = "LOGIN"
)

var SMTPAuths = []string{SMTP_PLAIN, SMTP_LOGIN}

func SMTPAuth(a smtp.Auth, cfg *SMTPConfig) error {
	c, err := smtp.Dial(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	if err != nil {
		return err
	}
	defer c.Close()

	if err = c.Hello("gogs"); err != nil {
		return err
	}

	if cfg.TLS {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err = c.StartTLS(&tls.Config{
				InsecureSkipVerify: cfg.SkipVerify,
				ServerName:         cfg.Host,
			}); err != nil {
				return err
			}
		} else {
			return errors.New("SMTP server unsupports TLS")
		}
	}

	if ok, _ := c.Extension("AUTH"); ok {
		if err = c.Auth(a); err != nil {
			return err
		}
		return nil
	}
	return errors.New("Unsupported SMTP authentication method")
}

// LoginViaSMTP queries if login/password is valid against the SMTP,
// and create a local user if success when enabled.
func LoginViaSMTP(user *User, login, password string, sourceID int64, cfg *SMTPConfig, autoRegister bool) (*User, error) {
	// Verify allowed domains.
	if len(cfg.AllowedDomains) > 0 {
		idx := strings.Index(login, "@")
		if idx == -1 {
			return nil, errors.UserNotExist{0, login}
		} else if !com.IsSliceContainsStr(strings.Split(cfg.AllowedDomains, ","), login[idx+1:]) {
			return nil, errors.UserNotExist{0, login}
		}
	}

	var auth smtp.Auth
	if cfg.Auth == SMTP_PLAIN {
		auth = smtp.PlainAuth("", login, password, cfg.Host)
	} else if cfg.Auth == SMTP_LOGIN {
		auth = &smtpLoginAuth{login, password}
	} else {
		return nil, errors.New("Unsupported SMTP authentication type")
	}

	if err := SMTPAuth(auth, cfg); err != nil {
		// Check standard error format first,
		// then fallback to worse case.
		tperr, ok := err.(*textproto.Error)
		if (ok && tperr.Code == 535) ||
			strings.Contains(err.Error(), "Username and Password not accepted") {
			return nil, errors.UserNotExist{0, login}
		}
		return nil, err
	}

	if !autoRegister {
		return user, nil
	}

	username := login
	idx := strings.Index(login, "@")
	if idx > -1 {
		username = login[:idx]
	}

	user = &User{
		LowerName:   strings.ToLower(username),
		Name:        strings.ToLower(username),
		Email:       login,
		Passwd:      password,
		LoginType:   LOGIN_SMTP,
		LoginSource: sourceID,
		LoginName:   login,
		IsActive:    true,
	}
	return user, CreateUser(user)
}

// __________  _____      _____
// \______   \/  _  \    /     \
//  |     ___/  /_\  \  /  \ /  \
//  |    |  /    |    \/    Y    \
//  |____|  \____|__  /\____|__  /
//                  \/         \/

// LoginViaPAM queries if login/password is valid against the PAM,
// and create a local user if success when enabled.
func LoginViaPAM(user *User, login, password string, sourceID int64, cfg *PAMConfig, autoRegister bool) (*User, error) {
	if err := pam.PAMAuth(cfg.ServiceName, login, password); err != nil {
		if strings.Contains(err.Error(), "Authentication failure") {
			return nil, errors.UserNotExist{0, login}
		}
		return nil, err
	}

	if !autoRegister {
		return user, nil
	}

	user = &User{
		LowerName:   strings.ToLower(login),
		Name:        login,
		Email:       login,
		Passwd:      password,
		LoginType:   LOGIN_PAM,
		LoginSource: sourceID,
		LoginName:   login,
		IsActive:    true,
	}
	return user, CreateUser(user)
}

func ExternalUserLogin(user *User, login, password string, source *LoginSource, autoRegister bool) (*User, error) {
	if !source.IsActived {
		return nil, errors.LoginSourceNotActivated{source.ID}
	}

	switch source.Type {
	case LOGIN_LDAP, LOGIN_DLDAP:
		return LoginViaLDAP(user, login, password, source, autoRegister)
	case LOGIN_SMTP:
		return LoginViaSMTP(user, login, password, source.ID, source.Cfg.(*SMTPConfig), autoRegister)
	case LOGIN_PAM:
		return LoginViaPAM(user, login, password, source.ID, source.Cfg.(*PAMConfig), autoRegister)
	}

	return nil, errors.InvalidLoginSourceType{source.Type}
}

// UserSignIn validates user name and password.
func UserSignIn(username, password string) (*User, error) {
	var user *User
	if strings.Contains(username, "@") {
		user = &User{Email: strings.ToLower(username)}
	} else {
		user = &User{LowerName: strings.ToLower(username)}
	}

	hasUser, err := x.Get(user)
	if err != nil {
		return nil, err
	}

	if hasUser {
		switch user.LoginType {
		case LOGIN_NOTYPE, LOGIN_PLAIN:
			if user.ValidatePassword(password) {
				return user, nil
			}

			return nil, errors.UserNotExist{user.ID, user.Name}

		default:
			var source LoginSource
			hasSource, err := x.Id(user.LoginSource).Get(&source)
			if err != nil {
				return nil, err
			} else if !hasSource {
				return nil, ErrLoginSourceNotExist{user.LoginSource}
			}

			return ExternalUserLogin(user, user.LoginName, password, &source, false)
		}
	}

	sources := make([]*LoginSource, 0, 3)
	if err = x.UseBool().Find(&sources, &LoginSource{IsActived: true}); err != nil {
		return nil, err
	}

	for _, source := range sources {
		authUser, err := ExternalUserLogin(nil, username, password, source, true)
		if err == nil {
			return authUser, nil
		}

		log.Warn("Failed to login '%s' via '%s': %v", username, source.Name, err)
	}

	return nil, errors.UserNotExist{user.ID, user.Name}
}
