package x2go

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

type X2Go struct {
	dec *xml.Decoder
	pre *xml.Decoder
}

func New(b []byte) *X2Go {
	return &X2Go{
		dec: xml.NewDecoder(bytes.NewReader(b)),
		pre: xml.NewDecoder(bytes.NewReader(b)),
	}
}

type node struct {
	name  xml.Name
	_type string
}

func (x *X2Go) String() string {
	bones := x.Skeleton()
	id := Identify(bones)
	return echo(id)
}

func echo(id map[string]map[string]string) string {
	s := ""
	var names []string
	for k, v := range id {
		if k == "" {
			continue
		}

		fmt.Println("type", strings.Title(k), "struct {")
		s += "type " + strings.Title(k) + " struct {\n"
		for name, typ := range v {
			if strings.Contains(name, ",") {
				names = strings.Split(name, ",")
				names = strings.Split(names[0], ":")
			} else if strings.Contains(name, ":") {
				names = strings.Split(name, ":")
			}
			s += "    " + strings.Title(names[len(names)-1]) + " " + typ + " `xml:" + `"` + name + `"` + "`\n"
			fmt.Println(" ", strings.Title(names[len(names)-1]), typ, "`xml:"+`"`+name+`"`+"`")
		}
		s += "}\n\n"
		fmt.Println("}")
	}

	return s
}

func (x *X2Go) Layer() int {
	count := 0
	names := []string{}

	for token, err := x.dec.Token(); err == nil; token, err = x.dec.Token() {
		switch t := token.(type) {
		case xml.StartElement:
			names = append(names, t.Name.Local)
			if count < len(names) {
				count = len(names)
			}
		case xml.EndElement:
			if t.Name.Local == names[len(names)-1] {
				names = names[:len(names)-1]
			}
		}
	}

	return count - 1
}

func (x *X2Go) namespace() (map[string]string, map[string][]string) {
	ns := map[string]string{}
	attrs := map[string][]string{}

	for token, err := x.pre.Token(); err == nil; token, err = x.pre.Token() {
		switch t := token.(type) {
		case xml.StartElement:
			if len(t.Attr) != 0 {
				for _, v := range t.Attr {
					ns[v.Value] = v.Name.Local
					attrs[t.Name.Local] = append(attrs[t.Name.Local], v.Name.Space+":"+v.Name.Local+",attr")
				}
			}
		}
	}

	return ns, attrs
}

func (x *X2Go) Skeleton() map[string][]string {
	bones := map[string][]string{}
	names := []string{}
	mapping := map[string]string{}
	ns, attrs := x.namespace()

	var val = func(s []string) string {
		if len(s) > 1 {
			return s[len(s)-2]
		}
		return ""
	}

	for token, err := x.dec.Token(); err == nil; token, err = x.dec.Token() {
		switch t := token.(type) {
		case xml.StartElement:
			name := t.Name.Local

			if t.Name.Space != "" {
				if ns[t.Name.Space] != "" {
					name = ns[t.Name.Space] + ":" + name
				}
			}

			if attrs[t.Name.Local] != nil {
				attrs[name] = attrs[t.Name.Local]
			}

			names = append(names, name)

			key := names[len(names)-1]
			value := val(names)

			if mapping[key] != "" && mapping[key] != value {
				bones[mapping[key]] = append(bones[mapping[key]], key)
			}
			mapping[names[len(names)-1]] = val(names)

		case xml.EndElement:
			name := t.Name.Local
			if t.Name.Space != "" {
				name = ns[t.Name.Space] + ":" + name
			}
			if name == names[len(names)-1] {
				names = names[:len(names)-1]
			}
		}
	}

	for k, v := range mapping {
		bones[v] = append(bones[v], k)
	}

	for k, v := range bones {
		bones[k] = append(v, attrs[k]...)
	}
	return bones
}

func Identify(bone map[string][]string) map[string]map[string]string {
	id := map[string]map[string]string{}
	for k, v := range bone {
		child := map[string]string{}
		for i := range v {
			if bone[v[i]] == nil {
				child[v[i]] = "string"
			} else {
				child[v[i]] = strings.Title(v[i])
			}
		}
		ids := strings.Split(k, ":")
		id[ids[len(ids)-1]] = child
	}

	return id
}
