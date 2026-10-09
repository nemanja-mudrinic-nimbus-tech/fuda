package markdown

import (
	"bytes"
	"net/url"
	"path"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type Links struct {
	Prefix    string
	DocsRoot  string
	Docs      map[string]bool
	Assets    map[string]bool
	TaskPaths map[string]string
	CodeURL   func(repoPath string) string
}

var engine = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
)

func Render(source []byte, docPath string, links Links) (string, error) {
	doc := engine.Parser().Parse(text.NewReader(source))

	var found []*ast.Link
	var images []*ast.Image
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch node := n.(type) {
		case *ast.Link:
			found = append(found, node)
		case *ast.Image:
			images = append(images, node)
		}
		return ast.WalkContinue, nil
	})
	for _, image := range images {
		image.Destination = []byte(links.imageSource(string(image.Destination), path.Dir(docPath)))
	}
	for _, link := range found {
		target, keep := links.resolve(string(link.Destination), path.Dir(docPath))
		if !keep {
			unwrap(link)
			continue
		}
		link.Destination = []byte(target)
		if isExternal(target) {
			link.SetAttributeString("target", "_blank")
			link.SetAttributeString("rel", "noreferrer")
		}
	}

	var out bytes.Buffer
	if err := engine.Renderer().Render(&out, source, doc); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (l Links) resolve(dest, dir string) (string, bool) {
	if dest == "" || strings.HasPrefix(dest, "#") {
		return dest, true
	}
	if isExternal(dest) || strings.HasPrefix(dest, "mailto:") {
		return dest, true
	}
	if strings.HasPrefix(dest, "/guide/") {
		return dest, true
	}
	if strings.HasPrefix(dest, "/") {
		return "", false
	}

	target, fragment, _ := strings.Cut(dest, "#")
	repoPath := path.Clean(path.Join(dir, target))
	if decoded, err := url.PathUnescape(repoPath); err == nil {
		repoPath = decoded
	}

	if id, ok := l.TaskPaths[repoPath]; ok {
		return l.Prefix + "/?task=" + url.QueryEscape(id), true
	}
	if l.Assets[repoPath] {
		return l.assetURL(repoPath), true
	}
	if strings.HasSuffix(repoPath, ".md") {
		if !l.Docs[repoPath] {
			return "", false
		}
		href := l.Prefix + "/docs/" + strings.TrimPrefix(repoPath, l.DocsRoot+"/")
		if fragment != "" {
			href += "#" + fragment
		}
		return href, true
	}
	if l.CodeURL == nil || strings.HasPrefix(repoPath, "..") {
		return "", false
	}
	if href := l.CodeURL(repoPath); href != "" {
		return href, true
	}
	return "", false
}

func (l Links) imageSource(dest, dir string) string {
	if isExternal(dest) || strings.HasPrefix(dest, "data:") || strings.HasPrefix(dest, "/") {
		return dest
	}
	repoPath := l.repoPath(dest, dir)
	if l.Assets[repoPath] {
		return l.assetURL(repoPath)
	}
	return ""
}

func (l Links) repoPath(dest, dir string) string {
	target, _, _ := strings.Cut(dest, "#")
	repoPath := path.Clean(path.Join(dir, target))
	if decoded, err := url.PathUnescape(repoPath); err == nil {
		return decoded
	}
	return repoPath
}

func (l Links) assetURL(repoPath string) string {
	return "/api" + l.Prefix + "/files?path=" + url.QueryEscape(repoPath)
}

func isExternal(dest string) bool {
	return strings.HasPrefix(dest, "http://") || strings.HasPrefix(dest, "https://")
}

func unwrap(link *ast.Link) {
	parent := link.Parent()
	for child := link.FirstChild(); child != nil; {
		next := child.NextSibling()
		parent.InsertBefore(parent, link, child)
		child = next
	}
	parent.RemoveChild(parent, link)
}

func Title(source []byte, fallback string) string {
	for line := range strings.SplitSeq(string(source), "\n") {
		if heading, ok := strings.CutPrefix(line, "# "); ok {
			return strings.TrimSpace(heading)
		}
	}
	return fallback
}
