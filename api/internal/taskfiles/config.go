package taskfiles

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

func parseConfig(files map[string][]byte, layout Layout, problems *[]Problem) Config {
	var c Config
	load := func(name string, decode func(map[string]*yaml.Node) error) {
		p := path.Join(layout.BoardDir, name)
		content, ok := files[p]
		if !ok {
			return
		}
		if err := decodeConfig(content, decode); err != nil {
			*problems = append(*problems, Problem{Path: p, Reason: err.Error() + "; ignored, derived from the tasks instead"})
		}
	}
	load("stages.md", func(keys map[string]*yaml.Node) error {
		return decodeStages(required(keys, "stages"), &c.Stages)
	})
	load("labels.md", func(keys map[string]*yaml.Node) error {
		if err := decodeLabelGroups(required(keys, "groups"), &c.LabelGroups); err != nil {
			return err
		}
		if colors, ok := keys["colors"]; ok {
			return colors.Decode(&c.LabelColors)
		}
		return nil
	})
	load("people.md", func(keys map[string]*yaml.Node) error {
		return decodePeople(required(keys, "people"), &c.People)
	})
	load("repos.md", func(keys map[string]*yaml.Node) error {
		return decodeCodeRepos(required(keys, "code_repos"), &c.CodeRepos)
	})
	return c
}

var missingKey = &yaml.Node{}

func required(keys map[string]*yaml.Node, key string) *yaml.Node {
	if n, ok := keys[key]; ok {
		return n
	}
	return missingKey
}

func decodeConfig(content []byte, decode func(map[string]*yaml.Node) error) error {
	front, _, err := splitFrontmatter(content)
	if err != nil {
		return err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(front, &doc); err != nil {
		return errors.New("invalid YAML: " + err.Error())
	}
	pairs, err := mappingPairs(&doc)
	if err != nil {
		return err
	}
	keys := map[string]*yaml.Node{}
	for _, p := range pairs {
		keys[p.key] = p.value
	}
	return decode(keys)
}

type stageEntry struct {
	Name   string   `yaml:"name"`
	Status []string `yaml:"status"`
	When   string   `yaml:"when"`
}

func decodeStages(n *yaml.Node, dst *[]Stage) error {
	if n == missingKey {
		return errors.New("stages: missing the stages key")
	}
	var entries []stageEntry
	if err := n.Decode(&entries); err != nil {
		return fmt.Errorf("stages: %w", err)
	}
	if len(entries) == 0 {
		return errors.New("stages: the list is empty")
	}
	stages := make([]Stage, 0, len(entries))
	for i, e := range entries {
		if e.Name == "" {
			return fmt.Errorf("stages: entry %d has no name", i+1)
		}
		if e.When != "" && e.When != "pr-open" {
			return fmt.Errorf("stages: %q has unknown when %q (only pr-open)", e.Name, e.When)
		}
		if len(e.Status) == 0 && e.When == "" {
			return fmt.Errorf("stages: %q needs a status list or when: pr-open", e.Name)
		}
		stages = append(stages, Stage{Name: e.Name, Statuses: e.Status, WhenPROpen: e.When == "pr-open"})
	}
	*dst = stages
	return nil
}

func decodeLabelGroups(n *yaml.Node, dst *[]LabelGroup) error {
	if n == missingKey {
		return errors.New("groups: missing the groups key")
	}
	if n.Kind != yaml.MappingNode {
		return errors.New("groups: expected a mapping of group names to value lists")
	}
	groups := make([]LabelGroup, 0, len(n.Content)/2)
	for i := 0; i+1 < len(n.Content); i += 2 {
		values, err := stringList(n.Content[i+1])
		if err != nil {
			return fmt.Errorf("groups: %q: %w", n.Content[i].Value, err)
		}
		groups = append(groups, LabelGroup{Name: n.Content[i].Value, Values: values})
	}
	*dst = groups
	return nil
}

type personEntry struct {
	Name    string   `yaml:"name"`
	Aliases []string `yaml:"aliases"`
}

func decodePeople(n *yaml.Node, dst *[]Person) error {
	if n == missingKey {
		return errors.New("people: missing the people key")
	}
	var entries []personEntry
	if err := n.Decode(&entries); err != nil {
		return fmt.Errorf("people: %w", err)
	}
	people := make([]Person, 0, len(entries))
	for i, e := range entries {
		if e.Name == "" {
			return fmt.Errorf("people: entry %d has no name", i+1)
		}
		people = append(people, Person(e))
	}
	*dst = people
	return nil
}

var codeRepo = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*(/[A-Za-z0-9][A-Za-z0-9._-]*){1,2}$`)

func decodeCodeRepos(n *yaml.Node, dst *[]string) error {
	if n == missingKey {
		return errors.New("code_repos: missing the code_repos key")
	}
	repos, err := stringList(n)
	if err != nil {
		return fmt.Errorf("code_repos: %w", err)
	}
	for _, repo := range repos {
		if !codeRepo.MatchString(repo) || strings.Contains(repo, "..") {
			return fmt.Errorf("code_repos: %q is not a repository like org/app", repo)
		}
	}
	*dst = repos
	return nil
}
