// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/json-iterator/go"
	_ "github.com/lib/pq"
	"github.com/unknwon/com"
	log "unknwon.dev/clog/v2"
	"xorm.io/core"
	"xorm.io/xorm"

	"github.com/G-Node/gogs/internal/conf"
	"github.com/G-Node/gogs/internal/db/migrations"
)

// Engine represents a XORM engine or session.
type Engine interface {
	Delete(interface{}) (int64, error)
	Exec(...interface{}) (sql.Result, error)
	Find(interface{}, ...interface{}) error
	Get(interface{}) (bool, error)
	ID(interface{}) *xorm.Session
	In(string, ...interface{}) *xorm.Session
	Insert(...interface{}) (int64, error)
	InsertOne(interface{}) (int64, error)
	Iterate(interface{}, xorm.IterFunc) error
	Sql(string, ...interface{}) *xorm.Session
	Table(interface{}) *xorm.Session
	Where(interface{}, ...interface{}) *xorm.Session
}

var (
	x         *xorm.Engine
	tables    []interface{}
	HasEngine bool

	EnableSQLite3 bool
)

func init() {
	tables = append(tables,
		new(User), new(PublicKey), new(AccessToken), new(TwoFactor), new(TwoFactorRecoveryCode),
		new(Repository), new(DeployKey), new(Collaboration), new(Access), new(Upload),
		new(Watch), new(Star), new(Follow), new(Action),
		new(Issue), new(PullRequest), new(Comment), new(Attachment), new(IssueUser),
		new(Label), new(IssueLabel), new(Milestone),
		new(Mirror), new(Release), new(LoginSource), new(Webhook), new(HookTask),
		new(ProtectBranch), new(ProtectBranchWhitelist),
		new(Team), new(OrgUser), new(TeamUser), new(TeamRepo),
		new(Notice), new(EmailAddress))

	gonicNames := []string{"SSL"}
	for _, name := range gonicNames {
		core.LintGonicMapper[name] = true
	}
}

// parsePostgreSQLHostPort parses given input in various forms defined in
// https://www.postgresql.org/docs/current/static/libpq-connect.html#LIBPQ-CONNSTRING
// and returns proper host and port number.
func parsePostgreSQLHostPort(info string) (string, string) {
	host, port := "127.0.0.1", "5432"
	if strings.Contains(info, ":") && !strings.HasSuffix(info, "]") {
		idx := strings.LastIndex(info, ":")
		host = info[:idx]
		port = info[idx+1:]
	} else if len(info) > 0 {
		host = info
	}
	return host, port
}

func parseMSSQLHostPort(info string) (string, string) {
	host, port := "127.0.0.1", "1433"
	if strings.Contains(info, ":") {
		host = strings.Split(info, ":")[0]
		port = strings.Split(info, ":")[1]
	} else if strings.Contains(info, ",") {
		host = strings.Split(info, ",")[0]
		port = strings.TrimSpace(strings.Split(info, ",")[1])
	} else if len(info) > 0 {
		host = info
	}
	return host, port
}

func getEngine() (*xorm.Engine, error) {
	Param := "?"
	if strings.Contains(conf.Database.Name, Param) {
		Param = "&"
	}

	connStr := ""
	switch conf.Database.Type {
	case "mysql":
		conf.UseMySQL = true
		if conf.Database.Host[0] == '/' { // looks like a unix socket
			connStr = fmt.Sprintf("%s:%s@unix(%s)/%s%scharset=utf8mb4&parseTime=true",
				conf.Database.User, conf.Database.Password, conf.Database.Host, conf.Database.Name, Param)
		} else {
			connStr = fmt.Sprintf("%s:%s@tcp(%s)/%s%scharset=utf8mb4&parseTime=true",
				conf.Database.User, conf.Database.Password, conf.Database.Host, conf.Database.Name, Param)
		}
		var engineParams = map[string]string{"rowFormat": "DYNAMIC"}
		return xorm.NewEngineWithParams(conf.Database.Type, connStr, engineParams)

	case "postgres":
		conf.UsePostgreSQL = true
		host, port := parsePostgreSQLHostPort(conf.Database.Host)
		if host[0] == '/' { // looks like a unix socket
			connStr = fmt.Sprintf("postgres://%s:%s@:%s/%s%ssslmode=%s&host=%s",
				url.QueryEscape(conf.Database.User), url.QueryEscape(conf.Database.Password), port, conf.Database.Name, Param, conf.Database.SSLMode, host)
		} else {
			connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s%ssslmode=%s",
				url.QueryEscape(conf.Database.User), url.QueryEscape(conf.Database.Password), host, port, conf.Database.Name, Param, conf.Database.SSLMode)
		}

	case "mssql":
		conf.UseMSSQL = true
		host, port := parseMSSQLHostPort(conf.Database.Host)
		connStr = fmt.Sprintf("server=%s; port=%s; database=%s; user id=%s; password=%s;", host, port, conf.Database.Name, conf.Database.User, conf.Database.Passwd)

	case "sqlite3":
		if !EnableSQLite3 {
			return nil, errors.New("this binary version does not build support for SQLite3")
		}
		if err := os.MkdirAll(path.Dir(conf.Database.Path), os.ModePerm); err != nil {
			return nil, fmt.Errorf("create directories: %v", err)
		}
		conf.UseSQLite3 = true
		connStr = "file:" + conf.Database.Path + "?cache=shared&mode=rwc"

	default:
		return nil, fmt.Errorf("unknown database type: %s", conf.Database.Type)
	}
	return xorm.NewEngine(conf.Database.Type, connStr)
}

func NewTestEngine(x *xorm.Engine) (err error) {
	x, err = getEngine()
	if err != nil {
		return fmt.Errorf("connect to database: %v", err)
	}

	x.SetMapper(core.GonicMapper{})
	return x.StoreEngine("InnoDB").Sync2(tables...)
}

func SetEngine() (err error) {
	x, err = getEngine()
	if err != nil {
		return fmt.Errorf("connect to database: %v", err)
	}

	x.SetMapper(core.GonicMapper{})

	// WARNING: for serv command, MUST remove the output to os.stdout,
	// so use log file to instead print to stdout.
	sec := conf.File.Section("log.xorm")
	logger, err := log.NewFileWriter(path.Join(conf.Log.RootPath, "xorm.log"),
		log.FileRotationConfig{
			Rotate:  sec.Key("ROTATE").MustBool(true),
			Daily:   sec.Key("ROTATE_DAILY").MustBool(true),
			MaxSize: sec.Key("MAX_SIZE").MustInt64(100) * 1024 * 1024,
			MaxDays: sec.Key("MAX_DAYS").MustInt64(3),
		})
	if err != nil {
		return fmt.Errorf("create 'xorm.log': %v", err)
	}

	// To prevent mystery "MySQL: invalid connection" error,
	// see https://gogs.io/gogs/issues/5532.
	x.SetMaxIdleConns(0)
	x.SetConnMaxLifetime(time.Second)

	if conf.IsProdMode() {
		x.SetLogger(xorm.NewSimpleLogger3(logger, xorm.DEFAULT_LOG_PREFIX, xorm.DEFAULT_LOG_FLAG, core.LOG_WARNING))
	} else {
		x.SetLogger(xorm.NewSimpleLogger(logger))
	}
	x.ShowSQL(true)
	return nil
}

func NewEngine() (err error) {
	if err = SetEngine(); err != nil {
		return err
	}

	if err = migrations.Migrate(x); err != nil {
		return fmt.Errorf("migrate: %v", err)
	}

	if err = x.StoreEngine("InnoDB").Sync2(tables...); err != nil {
		return fmt.Errorf("sync structs to database tables: %v\n", err)
	}

	return nil
}

type Statistic struct {
	Counter struct {
		User, Org, PublicKey,
		Repo, Watch, Star, Action, Access,
		Issue, Comment, Oauth, Follow,
		Mirror, Release, LoginSource, Webhook,
		Milestone, Label, HookTask,
		Team, UpdateTask, Attachment int64
	}
}

func GetStatistic() (stats Statistic) {
	stats.Counter.User = CountUsers()
	stats.Counter.Org = CountOrganizations()
	stats.Counter.PublicKey, _ = x.Count(new(PublicKey))
	stats.Counter.Repo = CountRepositories(true)
	stats.Counter.Watch, _ = x.Count(new(Watch))
	stats.Counter.Star, _ = x.Count(new(Star))
	stats.Counter.Action, _ = x.Count(new(Action))
	stats.Counter.Access, _ = x.Count(new(Access))
	stats.Counter.Issue, _ = x.Count(new(Issue))
	stats.Counter.Comment, _ = x.Count(new(Comment))
	stats.Counter.Oauth = 0
	stats.Counter.Follow, _ = x.Count(new(Follow))
	stats.Counter.Mirror, _ = x.Count(new(Mirror))
	stats.Counter.Release, _ = x.Count(new(Release))
	stats.Counter.LoginSource = CountLoginSources()
	stats.Counter.Webhook, _ = x.Count(new(Webhook))
	stats.Counter.Milestone, _ = x.Count(new(Milestone))
	stats.Counter.Label, _ = x.Count(new(Label))
	stats.Counter.HookTask, _ = x.Count(new(HookTask))
	stats.Counter.Team, _ = x.Count(new(Team))
	stats.Counter.Attachment, _ = x.Count(new(Attachment))
	return
}

func Ping() error {
	return x.Ping()
}

// The version table. Should have only one row with id==1
type Version struct {
	ID      int64
	Version int64
}

// DumpDatabase dumps all data from database to file system in JSON format.
func DumpDatabase(dirPath string) (err error) {
	os.MkdirAll(dirPath, os.ModePerm)
	// Purposely create a local variable to not modify global variable
	tables := append(tables, new(Version))
	for _, table := range tables {
		tableName := strings.TrimPrefix(fmt.Sprintf("%T", table), "*db.")
		tableFile := path.Join(dirPath, tableName+".json")
		f, err := os.Create(tableFile)
		if err != nil {
			return fmt.Errorf("create JSON file: %v", err)
		}

		if err = x.Asc("id").Iterate(table, func(idx int, bean interface{}) (err error) {
			return jsoniter.NewEncoder(f).Encode(bean)
		}); err != nil {
			f.Close()
			return fmt.Errorf("dump table '%s': %v", tableName, err)
		}
		f.Close()
	}
	return nil
}

// ImportDatabase imports data from backup archive.
func ImportDatabase(dirPath string, verbose bool) (err error) {
	snakeMapper := core.SnakeMapper{}

	skipInsertProcessors := map[string]bool{
		"mirror":    true,
		"milestone": true,
	}

	// Purposely create a local variable to not modify global variable
	tables := append(tables, new(Version))
	for _, table := range tables {
		tableName := strings.TrimPrefix(fmt.Sprintf("%T", table), "*db.")
		tableFile := path.Join(dirPath, tableName+".json")
		if !com.IsExist(tableFile) {
			continue
		}

		if verbose {
			log.Trace("Importing table '%s'...", tableName)
		}

		if err = x.DropTables(table); err != nil {
			return fmt.Errorf("drop table '%s': %v", tableName, err)
		} else if err = x.Sync2(table); err != nil {
			return fmt.Errorf("sync table '%s': %v", tableName, err)
		}

		f, err := os.Open(tableFile)
		if err != nil {
			return fmt.Errorf("open JSON file: %v", err)
		}
		rawTableName := x.TableName(table)
		_, isInsertProcessor := table.(xorm.BeforeInsertProcessor)
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			switch bean := table.(type) {
			case *LoginSource:
				meta := make(map[string]interface{})
				if err = jsoniter.Unmarshal(scanner.Bytes(), &meta); err != nil {
					return fmt.Errorf("unmarshal to map: %v", err)
				}

				tp := LoginType(com.StrTo(com.ToStr(meta["Type"])).MustInt64())
				switch tp {
				case LOGIN_LDAP, LOGIN_DLDAP:
					bean.Cfg = new(LDAPConfig)
				case LOGIN_SMTP:
					bean.Cfg = new(SMTPConfig)
				case LOGIN_PAM:
					bean.Cfg = new(PAMConfig)
				case LOGIN_GITHUB:
					bean.Cfg = new(GitHubConfig)
				default:
					return fmt.Errorf("unrecognized login source type:: %v", tp)
				}
				table = bean
			}

			if err = jsoniter.Unmarshal(scanner.Bytes(), table); err != nil {
				return fmt.Errorf("unmarshal to struct: %v", err)
			}

			if _, err = x.Insert(table); err != nil {
				return fmt.Errorf("insert strcut: %v", err)
			}

			meta := make(map[string]interface{})
			if err = jsoniter.Unmarshal(scanner.Bytes(), &meta); err != nil {
				log.Error("Failed to unmarshal to map: %v", err)
			}

			// Reset created_unix back to the date save in archive because Insert method updates its value
			if isInsertProcessor && !skipInsertProcessors[rawTableName] {
				if _, err = x.Exec("UPDATE "+rawTableName+" SET created_unix=? WHERE id=?", meta["CreatedUnix"], meta["ID"]); err != nil {
					log.Error("Failed to reset 'created_unix': %v", err)
				}
			}

			switch rawTableName {
			case "milestone":
				if _, err = x.Exec("UPDATE "+rawTableName+" SET deadline_unix=?, closed_date_unix=? WHERE id=?", meta["DeadlineUnix"], meta["ClosedDateUnix"], meta["ID"]); err != nil {
					log.Error("Failed to reset 'milestone.deadline_unix', 'milestone.closed_date_unix': %v", err)
				}
			}
		}

		// PostgreSQL needs manually reset table sequence for auto increment keys
		if conf.UsePostgreSQL {
			rawTableName := snakeMapper.Obj2Table(tableName)
			seqName := rawTableName + "_id_seq"
			if _, err = x.Exec(fmt.Sprintf(`SELECT setval('%s', COALESCE((SELECT MAX(id)+1 FROM "%s"), 1), false);`, seqName, rawTableName)); err != nil {
				return fmt.Errorf("reset table '%s' sequence: %v", rawTableName, err)
			}
		}
	}
	return nil
}
