package gig

import (
	"io"
	"bytes"
	"strings"
	"compress/zlib"
)

func readUntilNul(r io.Reader) (string, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	for {
		var b [1]byte
		_, err := r.Read(b[:])
		if err != nil {
			return "", err
		} else if b[0] == 0 {
			break
		}
		buf.WriteByte(b[0])
	}

	return buf.String(), nil
}

func split2(s, sep string) (head, tail string) {
	comps := strings.SplitN(s, sep, 2)
	head = comps[0]
	if len(comps) > 1 {
		tail = comps[1]
	}
	return
}

type zlibReadCloser struct {
	io.LimitedReader     //R of io.LimitedReader is the zlib reader
	source io.ReadCloser //the underlying source
}

func (r *zlibReadCloser) Close() error {
	var e1, e2 error

	// this shouldn't fail ever actually, since the wrapped
	//  object should have been an io.ReadCloser
	if rc, ok := r.LimitedReader.R.(io.Closer); ok {
		e1 = rc.Close()
	}

	e2 = r.source.Close()

	if e1 == nil && e2 == nil {
		return nil
	} else if e2 != nil {
		return e2
	}
	return e1
}

func (o *gitObject) wrapSourceWithDeflate() error {
	r, err := zlib.NewReader(o.source)
	if err != nil {
		return err
	}

	o.source = &zlibReadCloser{io.LimitedReader{R: r, N: o.size}, o.source}
	return nil
}

func (o *gitObject) wrapSource(rc io.ReadCloser) {
	o.source = &zlibReadCloser{io.LimitedReader{R: rc, N: o.size}, o.source}
}
