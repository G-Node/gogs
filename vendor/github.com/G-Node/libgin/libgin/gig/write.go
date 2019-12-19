package gig

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func writeHeader(o Object, w *bufio.Writer) (n int64, err error) {

	x, err := w.WriteString(o.Type().String())
	n += int64(x)
	if err != nil {
		return n, err
	}

	x, err = w.WriteString(" ")
	n += int64(x)
	if err != nil {
		return n, err
	}

	x, err = w.WriteString(fmt.Sprintf("%d", o.Size()))
	n += int64(x)
	if err != nil {
		return n, err
	}

	err = w.WriteByte(0)
	if err != nil {
		return n, err
	}

	return n + 1, nil
}

//WriteTo writes the commit object to the writer in the on-disk format
//i.e. as it would be stored in the git objects dir (although uncompressed).
func (c *Commit) WriteTo(writer io.Writer) (int64, error) {
	w := bufio.NewWriter(writer)

	n, err := writeHeader(c, w)
	if err != nil {
		return n, err
	}

	x, err := w.WriteString(fmt.Sprintf("tree %s\n", c.Tree))
	n += int64(x)
	if err != nil {
		return n, err
	}

	for _, p := range c.Parent {
		x, err = w.WriteString(fmt.Sprintf("parent %s\n", p))
		n += int64(x)
		if err != nil {
			return n, err
		}
	}

	x, err = w.WriteString(fmt.Sprintf("author %s\n", c.Author))
	n += int64(x)
	if err != nil {
		return n, err
	}

	x, err = w.WriteString(fmt.Sprintf("committer %s\n", c.Committer))
	n += int64(x)
	if err != nil {
		return n, err
	}

	if c.GPGSig != "" {
		s := strings.Replace(c.GPGSig, "\n", "\n ", -1)
		x, err = w.WriteString(fmt.Sprintf("gpgsig %s\n", s))
		n += int64(x)
		if err != nil {
			return n, err
		}

	}

	x, err = w.WriteString(fmt.Sprintf("\n%s", c.Message))
	n += int64(x)
	if err != nil {
		return n, err
	}

	err = w.Flush()
	return n, err
}

//WriteTo writes the tree object to the writer in the on-disk format
//i.e. as it would be stored in the git objects dir (although uncompressed).
func (t *Tree) WriteTo(writer io.Writer) (int64, error) {

	w := bufio.NewWriter(writer)

	n, err := writeHeader(t, w)
	if err != nil {
		return n, err
	}

	for t.Next() {
		//format is: [mode{ASCII, octal}][space][name][\0][SHA1]
		entry := t.Entry()
		line := fmt.Sprintf("%o %s", entry.Mode, entry.Name)
		x, err := w.WriteString(line)
		n += int64(x)
		if err != nil {
			return n, err
		}

		err = w.WriteByte(0)
		if err != nil {
			return n, err
		}
		n++

		x, err = w.Write(entry.ID[:])
		n += int64(x)
		if err != nil {
			return n, err
		}
	}

	if err = t.Err(); err != nil {
		return n, err
	}

	err = w.Flush()
	return n, err
}

//WriteTo writes the blob object to the writer in the on-disk format
//i.e. as it would be stored in the git objects dir (although uncompressed).
func (b *Blob) WriteTo(writer io.Writer) (int64, error) {
	w := bufio.NewWriter(writer)

	n, err := writeHeader(b, w)
	if err != nil {
		return n, err
	}

	x, err := io.Copy(w, b.source)
	n += int64(x)
	if err != nil {
		return n, err
	}

	err = w.Flush()
	return n, err
}

//WriteTo writes the tag object to the writer in the on-disk format
//i.e. as it would be stored in the git objects dir (although uncompressed).
func (t *Tag) WriteTo(writer io.Writer) (int64, error) {
	w := bufio.NewWriter(writer)

	n, err := writeHeader(t, w)
	if err != nil {
		return n, err
	}

	x, err := w.WriteString(fmt.Sprintf("object %s\n", t.Object))
	n += int64(x)
	if err != nil {
		return n, err
	}

	x, err = w.WriteString(fmt.Sprintf("type %s\n", t.ObjType))
	n += int64(x)
	if err != nil {
		return n, err
	}

	x, err = w.WriteString(fmt.Sprintf("tag %s\n", t.Tag))
	n += int64(x)
	if err != nil {
		return n, err
	}

	x, err = w.WriteString(fmt.Sprintf("tagger %s\n\n", t.Tagger))
	n += int64(x)
	if err != nil {
		return n, err
	}

	x, err = w.WriteString(t.Message)
	n += int64(x)
	if err != nil {
		return n, err
	}
	if t.GPGSig != "" {
		x, err = w.WriteString(fmt.Sprintf("%s\n", t.GPGSig))
		n += int64(x)
		if err != nil {
			return n, err
		}

	}

	err = w.Flush()
	return n, err
}
