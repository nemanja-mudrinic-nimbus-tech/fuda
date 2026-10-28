package taskfiles

import (
	"regexp"
	"slices"
	"strings"
)

type Activity struct {
	When   string `json:"when,omitempty"`
	Actor  string `json:"actor,omitempty"`
	Event  string `json:"event,omitempty"`
	Result string `json:"result,omitempty"`
	Usage  string `json:"usage,omitempty"`
	Text   string `json:"text"`
}

const activityHeading = "## activity"

var bullet = regexp.MustCompile(`^[-*+](\s|$)`)

var activityTime = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$`)

func SplitActivity(body string) (rest string, entries []Activity) {
	lines := strings.Split(body, "\n")
	start, end := activityBounds(lines)
	if start < 0 {
		return body, nil
	}
	for _, line := range lines[start+1 : end] {
		if entry, ok := parseActivity(line); ok {
			entries = append(entries, entry)
		}
	}
	kept := append(append([]string{}, lines[:start]...), lines[end:]...)
	return strings.Join(kept, "\n"), entries
}

func activityBounds(lines []string) (start, end int) {
	start = -1
	fenced := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fenced = !fenced
			continue
		}
		if fenced {
			continue
		}
		switch {
		case start < 0 && strings.EqualFold(strings.TrimRight(trimmed, " #"), activityHeading):
			start = i
		case start >= 0 && endsSection(trimmed):
			return start, i
		}
	}
	return start, len(lines)
}

func endsSection(line string) bool {
	return strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") || line == "#" || line == "##"
}

func parseActivity(line string) (Activity, bool) {
	text := strings.TrimSpace(line)
	if bullet.MatchString(text) {
		text = strings.TrimSpace(text[1:])
	}
	if text == "" {
		return Activity{}, false
	}
	parts := strings.Split(text, "·")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if len(parts) < 4 || len(parts) > 5 || !activityTime.MatchString(parts[0]) || slices.Contains(parts[1:4], "") {
		return Activity{Text: text}, true
	}
	entry := Activity{When: parts[0], Actor: parts[1], Event: parts[2], Result: parts[3], Text: text}
	if len(parts) == 5 {
		entry.Usage = parts[4]
	}
	return entry, true
}
