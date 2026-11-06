package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	snowflake      = regexp.MustCompile(`^\d{17,20}$`)
	permissionKeys = map[string]bool{"permissions": true, "allow": true, "deny": true}
	secretKeys     = map[string]bool{"token": true, "session_id": true, "endpoint": true, "connection_id": true, "resume_gateway_url": true}
	imageKeys      = map[string]bool{"avatar": true, "banner": true, "icon": true, "splash": true}
	nameKeys       = map[string]bool{"username": true, "global_name": true}
	locationKeys   = map[string]json.RawMessage{"latitude": json.RawMessage(`"0"`), "longitude": json.RawMessage(`"0"`), "country_code": json.RawMessage(`"XX"`), "email": json.RawMessage(`null`), "phone": json.RawMessage(`"0"`), "ip": json.RawMessage(`"0"`)}
)

type Mapping struct {
	IDs   map[string]string `json:"ids"`
	Names map[string]string `json:"names"`
}

func NewMapping(keep []string) *Mapping {
	m := &Mapping{IDs: map[string]string{}, Names: map[string]string{}}
	for _, name := range keep {
		m.Names[name] = name
	}
	return m
}

func (m *Mapping) id(real string) string {
	if fake, ok := m.IDs[real]; ok {
		return fake
	}
	fake := strconv.FormatInt(100000000000000000+int64(len(m.IDs))+1, 10)
	m.IDs[real] = fake
	return fake
}

func (m *Mapping) name(real string) string {
	if fake, ok := m.Names[real]; ok {
		return fake
	}
	fake := fmt.Sprintf("user%d", len(m.Names))
	m.Names[real] = fake
	return fake
}

type member struct {
	key   string
	value node
}

type node struct {
	object []member
	array  []node
	isObj  bool
	isArr  bool
	scalar json.RawMessage
}

func parse(dec *json.Decoder) (node, error) {
	tok, err := dec.Token()
	if err != nil {
		return node{}, err
	}
	switch t := tok.(type) {
	case json.Delim:
		if t == '{' {
			n := node{isObj: true}
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return node{}, err
				}
				value, err := parse(dec)
				if err != nil {
					return node{}, err
				}
				n.object = append(n.object, member{keyTok.(string), value})
			}
			_, err = dec.Token()
			return n, err
		}
		n := node{isArr: true, array: []node{}}
		for dec.More() {
			value, err := parse(dec)
			if err != nil {
				return node{}, err
			}
			n.array = append(n.array, value)
		}
		_, err = dec.Token()
		return n, err
	case json.Number:
		return node{scalar: json.RawMessage(t.String())}, nil
	default:
		raw, err := json.Marshal(t)
		return node{scalar: raw}, err
	}
}

func (m *Mapping) walk(n node, key string) node {
	switch {
	case n.isObj:
		for _, field := range n.object {
			if field.key == "key" && field.value.scalar != nil {
				var name string
				if json.Unmarshal(field.value.scalar, &name) == nil && permissionKeys[name] {
					return n
				}
			}
		}
		out := node{isObj: true}
		for _, field := range n.object {
			if replacement, ok := locationKeys[field.key]; ok {
				out.object = append(out.object, member{field.key, node{scalar: replacement}})
				continue
			}
			k := field.key
			if snowflake.MatchString(k) {
				k = m.id(k)
			}
			out.object = append(out.object, member{k, m.walk(field.value, field.key)})
		}
		return out
	case n.isArr:
		out := node{isArr: true, array: make([]node, len(n.array))}
		for i, item := range n.array {
			out.array[i] = m.walk(item, key)
		}
		return out
	}
	var text string
	if len(n.scalar) == 0 || n.scalar[0] != '"' || json.Unmarshal(n.scalar, &text) != nil {
		return n
	}
	switch {
	case permissionKeys[key]:
		return n
	case key == "emoji_name":
		text = "emoji"
	case snowflake.MatchString(text):
		text = m.id(text)
	case nameKeys[key]:
		text = m.name(text)
	case imageKeys[key]:
		text = "hash"
	case key == "code":
		text = "invitecode"
	case secretKeys[key]:
		text = "redacted"
	default:
		return n
	}
	raw, _ := json.Marshal(text)
	return node{scalar: raw}
}

func (n node) plain() any {
	switch {
	case n.isObj:
		out := make(map[string]any, len(n.object))
		for _, field := range n.object {
			out[field.key] = field.value.plain()
		}
		return out
	case n.isArr:
		out := make([]any, len(n.array))
		for i, item := range n.array {
			out[i] = item.plain()
		}
		return out
	}
	return n.scalar
}

func (m *Mapping) Anonymize(r io.Reader) ([]byte, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	root, err := parse(dec)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m.walk(root, "").plain()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func run(in, out, mapFile string, keep []string) error {
	mapping := NewMapping(keep)
	if mapFile != "" {
		if data, err := os.ReadFile(mapFile); err == nil {
			if err := json.Unmarshal(data, mapping); err != nil {
				return fmt.Errorf("read %s: %w", mapFile, err)
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	dirs, err := os.ReadDir(in)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		files, _ := filepath.Glob(filepath.Join(in, dir.Name(), "*.json"))
		sort.Strings(files)
		for i, file := range files {
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			data, err := mapping.Anonymize(f)
			f.Close()
			if err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
			if err := os.WriteFile(filepath.Join(out, fmt.Sprintf("%s.%d.json", dir.Name(), i)), data, 0o644); err != nil {
				return err
			}
		}
	}
	if mapFile == "" {
		return nil
	}
	data, err := json.Marshal(mapping)
	if err != nil {
		return err
	}
	return os.WriteFile(mapFile, data, 0o600)
}

func main() {
	in := flag.String("in", "", "directory with one subdirectory of raw JSON payloads per event or response")
	out := flag.String("out", "", "directory for the anonymized EVENT.N.json files")
	mapFile := flag.String("map", "", "file that keeps the real-to-fake mapping between runs; never commit it")
	keep := flag.String("keep", "", "comma-separated user names to keep as they are")
	flag.Parse()
	if *in == "" || *out == "" {
		flag.Usage()
		os.Exit(2)
	}
	var names []string
	if *keep != "" {
		names = strings.Split(*keep, ",")
	}
	if err := run(*in, *out, *mapFile, names); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
