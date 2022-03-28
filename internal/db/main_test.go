// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	log "unknwon.dev/clog/v2"

	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/dbutil"
	"github.com/NII-DG/gogs/internal/testutil"
)

var printSQL = flag.Bool("print-sql", false, "Print SQL executed")

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		// Remove the primary logger and register a noop logger.
		log.Remove(log.DefaultConsoleName)
		err := log.New("noop", testutil.InitNoopLogger)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	now := time.Now().UTC().Truncate(time.Second)
	gorm.NowFunc = func() time.Time { return now }

	os.Exit(m.Run())
}

// clearTables removes all rows from given tables.
func clearTables(t *testing.T, db *gorm.DB, tables ...interface{}) error {
	if t.Failed() {
		return nil
	}

	for _, t := range tables {
		err := db.Delete(t).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func initTestDB(t *testing.T, suite string, tables ...interface{}) *gorm.DB {
	t.Helper()

	dbpath := filepath.Join(os.TempDir(), fmt.Sprintf("gogs-%s-%d.db", suite, time.Now().Unix()))
	db, err := openDB(conf.DatabaseOpts{
		Type: "sqlite3",
		Path: dbpath,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = db.Close()

		if t.Failed() {
			t.Logf("Database %q left intact for inspection", dbpath)
			return
		}

		_ = os.Remove(dbpath)
	})

	db.SingularTable(true)
	if !testing.Verbose() {
		db.SetLogger(&dbutil.Writer{Writer: ioutil.Discard})
	}
	if *printSQL {
		db.LogMode(true)
	}

	err = db.AutoMigrate(tables...).Error
	if err != nil {
		t.Fatal(err)
	}

	return db
}
