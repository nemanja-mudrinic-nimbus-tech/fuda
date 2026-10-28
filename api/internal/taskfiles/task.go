package taskfiles

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

var errNoFrontmatter = errors.New("no frontmatter: the file must start with a --- block")

func splitFrontmatter(content []byte) (front, body []byte, err error) {
	content = bytes.TrimPrefix(content, []byte("\xef\xbb\xbf"))
	content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return nil, nil, errNoFrontmatter
	}
	rest := content[len("---\n"):]
	if bytes.HasPrefix(rest, []byte("---\n")) || bytes.Equal(rest, []byte("---")) {
		return nil, bytes.TrimPrefix(rest[3:], []byte("\n")), nil
	}
	end := bytes.Index(rest, []byte("\n---\n"))
	if end < 0 {
		if bytes.HasSuffix(rest, []byte("\n---")) {
			return rest[:len(rest)-len("\n---")], nil, nil
		}
		return nil, nil, errors.New("frontmatter is not closed with ---")
	}
	return rest[:end], rest[end+len("\n---\n"):], nil
}

func ParseTask(p string, content []byte) (Task, error) {
	t, reason := parseTask(p, content, false)
	if reason != "" {
		return Task{}, errors.New(reason)
	}
	return t, nil
}

func parseTask(p string, content []byte, archived bool) (Task, string) {
	front, body, err := splitFrontmatter(content)
	if err != nil {
		return Task{}, err.Error()
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(front, &doc); err != nil {
		return Task{}, "invalid YAML: " + err.Error()
	}
	fields, err := mappingPairs(&doc)
	if err != nil {
		return Task{}, err.Error()
	}

	t := Task{Path: p, Body: string(body), Archived: archived}
	_, t.Activities = SplitActivity(t.Body)
	for _, f := range fields {
		if err := t.set(f.key, f.value); err != nil {
			return Task{}, fmt.Sprintf("field %q: %v", f.key, err)
		}
	}

	switch {
	case t.ID == "":
		return Task{}, "missing required field id"
	case t.Title == "":
		return Task{}, "missing required field title"
	case t.Status == "":
		return Task{}, "missing required field status"
	case !nameMatchesID(path.Base(p), t.ID):
		return Task{}, fmt.Sprintf("filename must start with the id %q", t.ID)
	case t.Status == "done" && !archived:
		return Task{}, `status "done" is only valid in the archive`
	}
	return t, ""
}

func nameMatchesID(name, id string) bool {
	stem := strings.TrimSuffix(name, ".md")
	return stem == id || strings.HasPrefix(stem, id+"-")
}

func (t *Task) set(key string, v *yaml.Node) error {
	switch key {
	case "id":
		return scalarInto(v, &t.ID)
	case "title":
		return scalarInto(v, &t.Title)
	case "status":
		return scalarInto(v, &t.Status)
	case "blocked_by":
		return scalarInto(v, &t.BlockedBy)
	case "owner":
		return peopleInto(v, &t.Owners)
	case "tester":
		return peopleInto(v, &t.Testers)
	case "labels":
		return labelsInto(v, &t.Labels)
	case "added":
		return scalarInto(v, &t.Added)
	case "claimed":
		return scalarInto(v, &t.Claimed)
	case "done":
		return scalarInto(v, &t.Done)
	case "pr":
		return prInto(v, t)
	}
	t.Custom = append(t.Custom, Field{Key: key, Value: display(v)})
	return nil
}

type pair struct {
	key   string
	value *yaml.Node
}

func mappingPairs(doc *yaml.Node) ([]pair, error) {
	if doc.Kind == 0 {
		return nil, nil
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, errors.New("frontmatter must be a mapping of fields")
	}
	m := doc.Content[0]
	pairs := make([]pair, 0, len(m.Content)/2)
	for i := 0; i+1 < len(m.Content); i += 2 {
		pairs = append(pairs, pair{key: m.Content[i].Value, value: m.Content[i+1]})
	}
	return pairs, nil
}

func isNull(v *yaml.Node) bool {
	return v.Kind == yaml.ScalarNode && v.Tag == "!!null"
}

func scalarInto(v *yaml.Node, dst *string) error {
	if isNull(v) {
		return nil
	}
	if v.Kind != yaml.ScalarNode {
		return errors.New("expected a single value")
	}
	*dst = strings.TrimSpace(v.Value)
	return nil
}

func stringList(v *yaml.Node) ([]string, error) {
	switch {
	case isNull(v):
		return nil, nil
	case v.Kind == yaml.ScalarNode:
		return []string{v.Value}, nil
	case v.Kind == yaml.SequenceNode:
		out := make([]string, 0, len(v.Content))
		for _, item := range v.Content {
			if item.Kind != yaml.ScalarNode {
				return nil, errors.New("expected a list of plain values")
			}
			out = append(out, item.Value)
		}
		return out, nil
	}
	return nil, errors.New("expected a value or a list")
}

func peopleInto(v *yaml.Node, dst *[]string) error {
	values, err := stringList(v)
	if err != nil {
		return err
	}
	for _, value := range values {
		for name := range strings.SplitSeq(value, ",") {
			name = strings.TrimSpace(name)
			if name != "" && name != "—" && name != "-" {
				*dst = append(*dst, name)
			}
		}
	}
	return nil
}

func labelsInto(v *yaml.Node, dst *[]Label) error {
	values, err := stringList(v)
	if err != nil {
		return err
	}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		group, val, found := strings.Cut(value, ":")
		if !found {
			*dst = append(*dst, Label{Value: value})
			continue
		}
		*dst = append(*dst, Label{Group: strings.TrimSpace(group), Value: strings.TrimSpace(val)})
	}
	return nil
}

func prInto(v *yaml.Node, t *Task) error {
	values, err := stringList(v)
	if err != nil {
		return err
	}
	var refs []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if n, err := strconv.Atoi(strings.TrimPrefix(value, "#")); err == nil {
			t.PRs = append(t.PRs, n)
			continue
		}
		refs = append(refs, value)
	}
	t.PRRef = strings.Join(refs, ", ")
	return nil
}

func display(v *yaml.Node) string {
	switch v.Kind {
	case yaml.ScalarNode:
		if isNull(v) {
			return ""
		}
		return v.Value
	case yaml.SequenceNode:
		parts := make([]string, 0, len(v.Content))
		for _, item := range v.Content {
			parts = append(parts, display(item))
		}
		return strings.Join(parts, ", ")
	}
	out, err := yaml.Marshal(v)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
