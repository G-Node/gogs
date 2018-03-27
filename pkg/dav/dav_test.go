package dav

import (
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
	fs := GinFS{"../../trepos"}
	f, err := fs.OpenFile("/user1/repo1/_dav/testfile1.txt", 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	st, err := f.Stat()
	if st.Name() != "testfile1.txt" {
		t.Fail()
		return
	}

	// lets try a directoty
	f, err = fs.OpenFile("/user1/repo1/_dav/", 0, 0)
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
	fs := GinFS{"../../trepos"}
	f, err := fs.OpenFile("/user1/repo1/_dav/", 0, 0)
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

	f, err = fs.OpenFile("/user1/repo1/_dav/fold1", 0, 0)
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
	fs := GinFS{"../../trepos"}
	f, err := fs.OpenFile("/user1/repo1/_dav/testfile1.txt", 0, 0)
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
	f, err = fs.OpenFile("/user1/repo1/_dav/fold1/file1.txt", 0, 0)
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
	fs := GinFS{"../../trepos"}
	f, err := fs.OpenFile("/user1/repo1/_dav/testfile1.txt", 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	stat, err := f.Stat()
	mtime := stat.ModTime().String()
	if (mtime!="2018-03-26 16:32:51 +0200 CEST"){
		t.Fail()
		return
	}
	return
}

func TestSeekFile(t *testing.T) {
	fs := GinFS{"../../trepos"}
	f, err := fs.OpenFile("/user1/repo1/_dav/testfile1.txt", 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	f.Seek(1,0)
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
}
