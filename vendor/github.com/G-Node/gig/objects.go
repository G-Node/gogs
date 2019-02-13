package gig

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

//SHA1 is the object identifying checksum of
// the object data
type SHA1 [20]byte

func (oid SHA1) String() string {
	return hex.EncodeToString(oid[:])
}

//ParseSHA1 expects a string with a hex encoded sha1.
//It will trim the string of newline and space before
//parsing.
func ParseSHA1(input string) (sha SHA1, err error) {
	data, err := hex.DecodeString(strings.Trim(input, " \n"))
	if err != nil {
		return
	} else if len(data) != 20 {
		err = fmt.Errorf("git: sha1 must be 20 bytes")
		return
	}

	copy(sha[:], data)
	return
}

//Signature is a combination of who (Name, Email) and when (Date, Offset).
//Used by Commit, Tag to link an action (committer, author, tagger, ...)
//with a person in a point in time.
type Signature struct {
	Name   string
	Email  string
	Date   time.Time
	Offset *time.Location
}

func (s Signature) String() string {
	return fmt.Sprintf("%s <%s> %d %s", s.Name, s.Email, s.Date.Unix(), s.Offset)
}

//ObjectType is to the git object type
type ObjectType byte

//The defined bits match the ones used in
//the git pack file format.
const (
	_         = iota
	ObjCommit = ObjectType(iota)
	ObjTree
	ObjBlob
	ObjTag

	ObjOFSDelta = ObjectType(0x6)
	ObjRefDelta = ObjectType(0x7)
)

//ParseObjectType takes a string and converts it
//to the corresponding ObjectType or error if
//the string doesn't match any type.
func ParseObjectType(s string) (ObjectType, error) {
	s = strings.Trim(s, "\n ")
	switch s {
	case "commit":
		return ObjCommit, nil
	case "tree":
		return ObjTree, nil
	case "blob":
		return ObjBlob, nil
	case "tag":
		return ObjTag, nil
	}

	return ObjectType(0), fmt.Errorf("git: unknown object: %q", s)
}

func (ot ObjectType) String() string {
	switch ot {
	case ObjCommit:
		return "commit"
	case ObjTree:
		return "tree"
	case ObjBlob:
		return "blob"
	case ObjTag:
		return "tag"
	case ObjOFSDelta:
		return "delta-ofs"
	case ObjRefDelta:
		return "delta-ref"
	}
	return "unknown"
}

//IsStandardObject checks if an object is
//one of the four common objects such as
//commit, tree, blob, tag.
func IsStandardObject(ot ObjectType) bool {
	return ot > 0 && ot < 5
}

//IsDeltaObject checks if an object is a
//delta object, i.e. OFSDelta or RefDelta
func IsDeltaObject(ot ObjectType) bool {
	return ot == ObjOFSDelta || ot == ObjRefDelta
}

//Object represents a git object. It has
//information common to all git objects,
//like their type and their size. Also,
//all git objects should be closed via
//Close().
type Object interface {
	Type() ObjectType
	Size() int64

	io.WriterTo
	io.Closer
}

type gitObject struct {
	otype ObjectType
	size  int64

	source io.ReadCloser
}

func (o *gitObject) Type() ObjectType {
	return o.otype
}

func (o *gitObject) Size() int64 {
	return o.size
}

func (o *gitObject) Close() error {
	if o.source == nil {
		return nil
	}
	return o.source.Close()
}

//Commit represents one git commit.
type Commit struct {
	gitObject

	Tree      SHA1
	Parent    []SHA1
	Author    Signature
	Committer Signature
	Message   string
	GPGSig    string
}

//Date returns the commit timestamps (with the correct location).
func (c *Commit) Date() time.Time {
	return c.Committer.Date.In(c.Committer.Offset)
}

//Tree represents the git tree object.
type Tree struct {
	gitObject

	entry *TreeEntry
	err   error
}

//TreeEntry holds information about a single
//entry in the git Tree object.
type TreeEntry struct {
	Mode os.FileMode
	Type ObjectType
	ID   SHA1
	Name string
}

//Next advances the pointer to the next TreeEntry
//within the Tree object. Returns false if it was
//pointing to the last element (EOF condition), or
//if there was an error while advacing. Use Err()
//to resolve between the to conditions.
func (tree *Tree) Next() bool {
	tree.entry, tree.err = parseTreeEntry(tree.source)
	return tree.err == nil
}

//Err returns the last error non-EOF error encountered.
func (tree *Tree) Err() error {
	if err := tree.err; err != nil && err != io.EOF {
		return err
	}

	return nil
}

//Entry returns the current TreeEntry.
func (tree *Tree) Entry() *TreeEntry {
	return tree.entry
}

//Blob represents a git blob object.
type Blob struct {
	gitObject
}

func (b *Blob) Read(data []byte) (n int, err error) {
	n, err = b.source.Read(data)
	return
}

func MakeAnnexBlob(fp *os.File, size int64) *Blob {
	return &Blob{gitObject{otype: ObjBlob, size: size, source: fp}}
}

//Tag represents a git tag object.
type Tag struct {
	gitObject

	Object  SHA1
	ObjType ObjectType
	Tag     string
	Tagger  Signature
	Message string
	GPGSig  string
}
