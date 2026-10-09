package taskfiles

import (
	"bytes"
	"errors"
	"strings"

	"go.yaml.in/yaml/v3"
)

var ErrNoStatusLine = errors.New("the frontmatter has no status line")

type Move struct {
	Status  string
	Claimed string
}

func ParseOne(p string, content []byte) (Task, error) {
	t, reason := parseTask(p, content, false)
	if reason != "" {
		return Task{}, errors.New(reason)
	}
	return t, nil
}

func ApplyMove(content []byte, m Move) ([]byte, error) {
	lines := bytes.SplitAfter(content, []byte("\n"))
	front := frontmatterLines(lines)
	statusAt, claimedAt := -1, -1
	for i := front.start; i < front.end; i++ {
		switch {
		case bytes.HasPrefix(lines[i], []byte("status:")):
			statusAt = i
		case bytes.HasPrefix(lines[i], []byte("claimed:")):
			claimedAt = i
		}
	}
	if statusAt < 0 {
		return nil, ErrNoStatusLine
	}

	eol := lineEnding(lines[statusAt])
	lines[statusAt] = []byte("status: " + yamlScalar(m.Status) + eol)
	if m.Claimed != "" {
		claimed := []byte("claimed: " + m.Claimed + eol)
		switch {
		case claimedAt < 0:
			lines = append(lines[:statusAt+1], append([][]byte{claimed}, lines[statusAt+1:]...)...)
		case strings.TrimSpace(string(lines[claimedAt][len("claimed:"):])) == "":
			lines[claimedAt] = claimed
		}
	}
	return bytes.Join(lines, nil), nil
}

func ApplyAssign(content []byte, owners []string) ([]byte, error) {
	lines := bytes.SplitAfter(content, []byte("\n"))
	front := frontmatterLines(lines)
	statusAt, ownerAt := -1, -1
	for i := front.start; i < front.end; i++ {
		switch {
		case bytes.HasPrefix(lines[i], []byte("status:")):
			statusAt = i
		case bytes.HasPrefix(lines[i], []byte("owner:")):
			ownerAt = i
		}
	}
	if statusAt < 0 {
		return nil, ErrNoStatusLine
	}

	if ownerAt < 0 {
		if len(owners) == 0 {
			return content, nil
		}
		line := []byte("owner: " + yamlScalar(strings.Join(owners, ", ")) + lineEnding(lines[statusAt]))
		lines = append(lines[:statusAt+1], append([][]byte{line}, lines[statusAt+1:]...)...)
		return bytes.Join(lines, nil), nil
	}

	itemsEnd := ownerAt + 1
	for itemsEnd < front.end && isListItem(lines[itemsEnd]) {
		itemsEnd++
	}
	var replacement [][]byte
	eol := lineEnding(lines[ownerAt])
	value := strings.TrimSpace(string(lines[ownerAt][len("owner:"):]))
	switch {
	case len(owners) == 0:
	case itemsEnd > ownerAt+1:
		replacement = append(replacement, lines[ownerAt])
		marker := listMarker(lines[ownerAt+1])
		for _, name := range owners {
			replacement = append(replacement, []byte(marker+yamlScalar(name)+eol))
		}
	case strings.HasPrefix(value, "["):
		quoted := make([]string, len(owners))
		for i, name := range owners {
			quoted[i] = yamlScalar(name)
		}
		replacement = append(replacement, []byte("owner: ["+strings.Join(quoted, ", ")+"]"+eol))
	default:
		replacement = append(replacement, []byte("owner: "+yamlScalar(strings.Join(owners, ", "))+eol))
	}
	out := append([][]byte{}, lines[:ownerAt]...)
	out = append(out, replacement...)
	out = append(out, lines[itemsEnd:]...)
	return bytes.Join(out, nil), nil
}

func isListItem(line []byte) bool {
	return bytes.HasPrefix(bytes.TrimLeft(line, " \t"), []byte("- "))
}

func listMarker(item []byte) string {
	trimmed := bytes.TrimLeft(item, " \t")
	return string(item[:len(item)-len(trimmed)]) + "- "
}

type frontmatterRange struct{ start, end int }

func frontmatterLines(lines [][]byte) frontmatterRange {
	if len(lines) == 0 || !isFence(bytes.TrimPrefix(lines[0], []byte("\xef\xbb\xbf"))) {
		return frontmatterRange{}
	}
	for i := 1; i < len(lines); i++ {
		if isFence(lines[i]) {
			return frontmatterRange{start: 1, end: i}
		}
	}
	return frontmatterRange{}
}

func isFence(line []byte) bool {
	return string(bytes.TrimRight(line, "\r\n")) == "---"
}

func lineEnding(line []byte) string {
	if bytes.HasSuffix(line, []byte("\r\n")) {
		return "\r\n"
	}
	return "\n"
}

func yamlScalar(s string) string {
	out, err := yaml.Marshal(s)
	if err != nil {
		return s
	}
	return strings.TrimSpace(string(out))
}
