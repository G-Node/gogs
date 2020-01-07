package markup

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"

	"golang.org/x/net/html/charset"
)

// ODML takes a []byte and returns a JSON []byte representation for rendering with jsTree
func MarshalODML(buf []byte) []byte {
	od := Odml{}
	decoder := xml.NewDecoder(bytes.NewReader(buf))
	decoder.CharsetReader = charset.NewReaderLabel
	decoder.Decode(&od)
	data, _ := json.Marshal(od)
	return data
}

type Section struct {
	Name       string       `json:"-" xml:"name"`
	Type       string       `json:"-" xml:"type"`
	Properties []Property   `json:"-" xml:"property"`
	Text       string       `json:"text"`
	Sections   []Section    `json:"-" xml:"section"`
	Children   []OdMLObject `json:"children,omitempty"`
}

type Property struct {
	Name       string   `json:"-" xml:"name"`
	Value      []string `json:"-" xml:"value"`
	Text       string   `json:"text"`
	Icon       string
	Definition string `xml:"definition"`
}

type OdMLObject struct {
	Prop    Property
	Section Section
	Type    string
}

type Odml struct {
	OdmnlSections []Section `json:"children" xml:"section"`
}

func (u *Property) MarshalJSON() ([]byte, error) {
	type Alias Property
	if u.Text == "" {
		u.Text = fmt.Sprintf("%s: %s (%s)", u.Name, u.Value, u.Definition)
	}
	return json.Marshal(Alias(*u))
}

func (u *Section) MarshalJSON() ([]byte, error) {
	type Alias Section
	if u.Text == "" {
		u.Text = fmt.Sprintf("%s", u.Name)
	}
	for _, x := range u.Properties {
		u.Children = append(u.Children, OdMLObject{Prop: x, Type: "property"})
	}
	for _, x := range u.Sections {
		u.Children = append(u.Children, OdMLObject{Section: x, Type: "section"})
	}
	return json.Marshal(Alias(*u))
}

func (u *OdMLObject) MarshalJSON() ([]byte, error) {
	if u.Type == "property" {
		return u.Prop.MarshalJSON()
	}
	if u.Type == "section" {
		return u.Section.MarshalJSON()
	}
	return nil, fmt.Errorf("Could not unmarshal odml object")
}
