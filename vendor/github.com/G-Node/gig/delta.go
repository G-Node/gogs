package gig

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

//Delta represents a git delta representation. Either BaseRef
//or BaseOff are valid fields, depending on its Type().
type Delta struct {
	gitObject

	BaseRef    SHA1
	BaseOff    int64
	SizeSource int64
	SizeTarget int64

	pf  *PackFile
	op  DeltaOp
	err error
}

func readByte(r io.Reader) (byte, error) {
	var err error
	n := 0
	var b [1]byte

	for n != 1 && err == nil {
		n, err = r.Read(b[:])
	}

	if n != 1 {
		return 0, err
	}

	return b[0], nil
}

func parseDelta(obj gitObject) (*Delta, error) {
	delta := Delta{gitObject: obj}

	//all delta objects come from a PackFile and
	//therefore git.Source is must be a *packReader
	source := delta.source.(*packReader)
	delta.pf = source.fd

	var err error
	if obj.otype == ObjRefDelta {
		_, err = source.Read(delta.BaseRef[:])
		//TODO: check n?

		if err != nil {
			return nil, err
		}

	} else {
		off, err := readVarint(source)
		if err != nil {
			return nil, err
		}

		delta.BaseOff = source.start - off
	}

	err = delta.wrapSourceWithDeflate()
	if err != nil {
		return nil, err
	}

	delta.SizeSource, err = readVarSize(delta.source, 0)
	if err != nil {
		return nil, err
	}

	delta.SizeTarget, err = readVarSize(delta.source, 0)
	if err != nil {
		return nil, err
	}

	return &delta, nil
}

func readVarSize(r io.Reader, offset uint) (size int64, err error) {
	size = int64(0)
	b := byte(0x80)

	// [0111 1111 ... 1111] (int64) is biggest decode-able
	// value we get by shifting byte b = 0x7F [0111 1111]
	// left 8*7 = 56 times; the next attempt must overflow.
	for i := offset; b&0x80 != 0 && i < 57; i += 7 {
		b, err = readByte(r)
		if err != nil {
			return 0, fmt.Errorf("git: io error: %v", err)
		}

		size |= int64(b&0x7F) << i
	}

	// means i > 56, would overflow (see above).
	if b&0x80 != 0 {
		return 0, fmt.Errorf("int64 overflow")
	}

	return size, nil
}

func decodeInt(r io.Reader, b byte, l uint) (size int64, err error) {

	for i := uint(0); i < l; i++ {

		if b&(1<<i) != 0 {
			var d byte
			d, err = readByte(r)
			if err != nil {
				return
			}

			size |= int64(d) << (i * 8)
		}
	}

	return
}

func readVarint(r io.Reader) (int64, error) {
	b, err := readByte(r)
	if err != nil {
		return 0, fmt.Errorf("git: io error: %v", err)
	}

	size := int64(b & 0x7F)

	for b&0x80 != 0 {
		b, err = readByte(r)
		if err != nil {
			return 0, fmt.Errorf("git: io error: %v", err)
		}

		size++

		// [0000 0001 ... 0000] (int64)
		//          ^ bit 0x38 (56)
		// shifting by 7 will shift the bit into the
		// sign bit of int64, i.e. we have overflow.
		if size > (1<<0x38)-1 {
			return 0, fmt.Errorf("int64 overflow")
		}

		size = (size << 7) + int64(b&0x7F)
	}

	return size, nil
}

//DeltaOpCode is the operation code for delta compression
//instruction set.
type DeltaOpCode byte

//DeltaOpCode values.
const (
	DeltaOpInsert = 1 //insert data from the delta data into dest
	DeltaOpCopy   = 2 //copy data from the original source into dest
)

//DeltaOp represents the delta compression operation. Offset is
//only valid for DeltaOpCopy operations.
type DeltaOp struct {
	Op     DeltaOpCode
	Size   int64
	Offset int64
}

//Op returns the current operations
func (d *Delta) Op() DeltaOp {
	return d.op
}

//Err retrieves the current error state, if any
func (d *Delta) Err() error {
	if err := d.err; err != io.EOF {
		return err
	}
	return nil
}

//NextOp reads the next DeltaOp from the delta data stream.
//Returns false when there are no operations left or on error;
//use Err() to decide between the two cases.
func (d *Delta) NextOp() bool {

	if d.err != nil {
		return false
	}

	b, err := readByte(d.source)
	if err != nil {
		return false
	}

	if b&0x80 != 0 {
		d.op.Op = DeltaOpCopy
		op := b & 0x7F
		d.op.Offset, d.err = decodeInt(d.source, op, 4)
		if d.err != nil {
			return false
		}

		d.op.Size, d.err = decodeInt(d.source, op>>4, 3)
		if d.err != nil {
			return false
		}

		if d.op.Size == 0 {
			d.op.Size = 0x10000
		}
	} else if n := b; n > 0 {
		d.op.Op = DeltaOpInsert
		d.op.Size = int64(n)
	} else {
		d.err = fmt.Errorf("git: unknown delta op code")
		return false
	}

	return true
}

//Patch applies the delta data onto r and writes the result to w.
func (d *Delta) Patch(r io.ReadSeeker, w io.Writer) error {

	for d.NextOp() {
		op := d.Op()
		switch op.Op {
		case DeltaOpCopy:
			_, err := r.Seek(op.Offset, os.SEEK_SET)
			if err != nil {
				return err
			}

			_, err = io.CopyN(w, r, op.Size)
			if err != nil {
				return err
			}
		case DeltaOpInsert:
			_, err := io.CopyN(w, d.source, op.Size)
			if err != nil {
				return err
			}
		}
	}

	return d.Err()
}

//SkipOp prepares the delta stream to move to the next operation
//without actually carrying out the delta operation. Useful for
//printing the delta stream.
func (d *Delta) SkipOp() {
	op := d.Op()
	if op.Op == DeltaOpInsert {
		_, d.err = io.CopyN(ioutil.Discard, d.source, op.Size)
	}
}

//WriteTo would write the object to disk in the git object
//representation. It is not NOT IMPLEMENTED for the delta
//object.
func (d *Delta) WriteTo(w io.Writer) (int64, error) {
	return 0, fmt.Errorf("WriteTo not implemented for Delta")
}

type deltaChain struct {
	baseObj gitObject
	baseOff int64

	links []Delta
}

func (c *deltaChain) Len() int {
	return len(c.links)
}

type objectSource interface {
	openRawObject(id SHA1) (gitObject, error)
}

func buildDeltaChain(d *Delta, s objectSource) (*deltaChain, error) {
	var chain deltaChain
	var err error

	for err == nil {

		chain.links = append(chain.links, *d)

		var obj gitObject
		if d.otype == ObjRefDelta {
			obj, err = s.openRawObject(d.BaseRef)
		} else {
			obj, err = d.pf.readRawObject(d.BaseOff)
		}

		if err != nil {
			break
		}

		if IsStandardObject(obj.otype) {
			chain.baseObj = obj
			chain.baseOff = d.BaseOff
			break
		} else if !IsDeltaObject(obj.otype) {
			err = fmt.Errorf("git: unexpected object type in delta chain")
			break
		}

		d, err = parseDelta(obj)
	}

	if err != nil {
		//cleanup
		return nil, err
	}

	return &chain, nil
}

func (c *deltaChain) resolve() (Object, error) {

	ibuf := bytes.NewBuffer(make([]byte, 0, c.baseObj.Size()))
	n, err := io.Copy(ibuf, c.baseObj.source)
	if err != nil {
		return nil, err
	}

	if n != c.baseObj.Size() {
		return nil, io.ErrUnexpectedEOF
	}

	obuf := bytes.NewBuffer(make([]byte, 0, c.baseObj.Size()))

	for i := len(c.links); i > 0; i-- {
		lk := c.links[i-1]

		if lk.SizeTarget > int64(^uint(0)>>1) {
			return nil, fmt.Errorf("git: target to large for delta unpatching")
		}

		obuf.Grow(int(lk.SizeTarget))
		obuf.Truncate(0)

		err = lk.Patch(bytes.NewReader(ibuf.Bytes()), obuf)

		if err != nil {
			return nil, err
		}

		if lk.SizeTarget != int64(obuf.Len()) {
			return nil, fmt.Errorf("git: size mismatch while patching delta object")
		}

		obuf, ibuf = ibuf, obuf
	}

	//ibuf is holding the data
	obj := gitObject{c.baseObj.otype, int64(ibuf.Len()), ioutil.NopCloser(ibuf)}
	return parseObject(obj)
}
