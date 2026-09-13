package utils

import (
	"html"
	"regexp"
	"strings"
)

var mdLinkRe = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
var mdBoldRe = regexp.MustCompile(`\*\*([^*]+)\*\*`)
var mdItalicRe = regexp.MustCompile(`\*([^*]+)\*`)
var mdCodeRe = regexp.MustCompile("`([^`]+)`")
var mdHRe = regexp.MustCompile(`(?m)^(#{1,6})\s*(.+)$`)
var mdListRe = regexp.MustCompile(`(?m)^[-*]\s+(.+)$`)

func RenderMarkdown(md string) string {
	md = html.EscapeString(md)
	md = mdHRe.ReplaceAllStringFunc(md, func(s string) string {
		m := mdHRe.FindStringSubmatch(s)
		level := len(m[1])
		return "<h" + itoa(level) + ">" + m[2] + "</h" + itoa(level) + ">"
	})
	md = mdListRe.ReplaceAllString(md, "<li>$1</li>")
	md = mdLinkRe.ReplaceAllString(md, `<a href="$2" target="_blank">$1</a>`)
	md = mdBoldRe.ReplaceAllString(md, "<strong>$1</strong>")
	md = mdItalicRe.ReplaceAllString(md, "<em>$1</em>")
	md = mdCodeRe.ReplaceAllString(md, "<code>$1</code>")
	lines := strings.Split(md, "\n")
	for i, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" && !strings.HasPrefix(l, "<") {
			lines[i] = "<p>" + l + "</p>"
		}
	}
	return strings.Join(lines, "\n")
}

func itoa(i int) string {
	switch i {
	case 1:
		return "1"
	case 2:
		return "2"
	case 3:
		return "3"
	case 4:
		return "4"
	case 5:
		return "5"
	case 6:
		return "6"
	}
	return "6"
}

func StripHTML(s string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	return strings.TrimSpace(re.ReplaceAllString(s, ""))
}

func Excerpt(md string, n int) string {
	text := StripHTML(RenderMarkdown(md))
	return Truncate(text, n)
}
