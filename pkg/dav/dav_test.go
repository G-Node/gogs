package dav

import (
	"io"
	"log"
	"strings"
	"testing"
)

func TestGetRepoName(t *testing.T) {
	name, err := getRName("/cgars/test/_dav/adasdasd/daasdas/asdasdsa")
	if err != nil {
		t.Logf("Repo Name not dtermined from path")
		t.Fail()
		return
	}
	if name != "test" {
		t.Logf("Repo Name not dtermined from path")
		t.Fail()
		return
	}
	return
}

func TestOwnerName(t *testing.T) {
	name, err := getOName("/cgars/test/_dav/adasdasd/daasdas/asdasdsa")
	if err != nil {
		t.Logf("Repo Name not dtermined from path")
		t.Fail()
		return
	}
	if name != "cgars" {
		t.Logf("Owner Name not dtermined from path: name was %s", name)
		t.Fail()
		return
	}
	return
}

func TestOpenfile(t *testing.T) {
	fs := GinFS{"../../testdata/trepos"}
	f, err := fs.OpenFile(nil, "/user1/repo1/_dav/testfile1.txt", 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	st, err := f.Stat()
	if st.Name() != "testfile1.txt" {
		t.Fail()
		return
	}

	// lets try a directoty
	f, err = fs.OpenFile(nil, "/user1/repo1/_dav/", 0, 0)
	st, err = f.Stat()
	if err != nil {
		log.Fatal(err)
	}
	if !st.IsDir() {
		t.Fail()
		return
	}
}

func TestReadDir(t *testing.T) {
	fs := GinFS{"../../testdata/trepos"}
	f, err := fs.OpenFile(nil, "/user1/repo1/_dav/", 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	ents, err := f.Readdir(0)
	if err != nil {
		log.Fatal(err)
	}
	if len(ents) < 1 {
		t.Log("Can not read directory")
		t.Fail()
		return
	}

	f, err = fs.OpenFile(nil, "/user1/repo1/_dav/fold1", 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	ents, err = f.Readdir(0)
	if err != nil {
		log.Fatal(err)
	}
	if len(ents) != 2 {
		t.Log("Can not read sub directory")
		t.Fail()
		return
	}

}

func TestReadFile(t *testing.T) {
	fs := GinFS{"../../testdata/trepos"}
	f, err := fs.OpenFile(nil, "/user1/repo1/_dav/testfile1.txt", 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	bf := make([]byte, 50)
	_, err = f.Read(bf)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	txt := string(bf)
	if !strings.Contains(txt, "test") {
		t.Log("could not read normal git file")
		t.Fail()
	}

	// Open a file in a subfolder
	f, err = fs.OpenFile(nil, "/user1/repo1/_dav/fold1/file1.txt", 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	bf = make([]byte, 50)
	_, err = f.Read(bf)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	txt = string(bf)
	if !strings.Contains(txt, "bla") {
		t.Log("could not read git file in sobfolder")
		t.Fail()
	}
}

func TestModFile(t *testing.T) {
	fs := GinFS{"../../testdata/trepos"}
	f, err := fs.OpenFile(nil, "/user1/repo1/_dav/testfile1.txt", 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	stat, err := f.Stat()
	mtime := stat.ModTime().String()
	if mtime != "2018-04-03 11:11:17 +0200 CEST" {
		t.Fail()
		return
	}
	return
}

func TestSeekFile(t *testing.T) {
	fs := GinFS{"../../testdata/trepos"}
	f, err := fs.OpenFile(nil, "/user1/repo1/_dav/testfile1.txt", 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	f.Seek(1, io.SeekStart)
	bf := make([]byte, 50)
	n, err := f.Read(bf)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	txt := string(bf)
	if n != 4 {
		t.Log("Read count wrong")
		t.Fail()
	}
	if !strings.Contains(txt, "est") {
		t.Log("could not read normal git file")
		t.Fail()
	}

	// Test Seek end Seek begin
	end, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		t.Logf("%*v", err)
		t.Fail()
	}
	beg, err := f.Seek(0, io.SeekStart)
	if err != nil {
		t.Logf("%*v", err)
		t.Fail()
	}
	if end-beg != 5 {
		t.Log("Seek end minus Seek begin is not size")
		t.Fail()
	}
}
