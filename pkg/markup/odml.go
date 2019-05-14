package markup

import (
	"bytes"
	"encoding/json"
	"encoding/xml"

	"github.com/G-Node/godML/odml"
	"golang.org/x/net/html/charset"
)

// ODML takes a []byte and returns a JSON []byte representation for rendering with jsTree
func MarshalODML(buf []byte) []byte {
	od := odml.Odml{}
	decoder := xml.NewDecoder(bytes.NewReader(buf))
	decoder.CharsetReader = charset.NewReaderLabel
	decoder.Decode(&od)
	data, _ := json.Marshal(od)
	return data
}
