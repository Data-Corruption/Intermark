package utils

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var md goldmark.Markdown

func InitMarkdownConverter() {
	md = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithInlineParsers(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
}

func MdToHTML(markdown string) (string, error) {
	out := markdown
	var mdsrc string
	var index int
	var err error = nil
	for {
		out, mdsrc, index, err = HTMLCutFirstNode(out, "mdsrc")
		if err != nil {
			break
		}
		// strip opening and closing tags
		mdsrc = mdsrc[len("<mdsrc>") : len(mdsrc)-len("</mdsrc>")]
		// convert to html
		var buf bytes.Buffer
		if err := md.Convert([]byte(mdsrc), &buf); err != nil {
			return "", err
		}
		// insert at index
		out = out[:index] + buf.String() + out[index:]
	}
	if err != ErrHTMLElementNotFound {
		return "", err
	}
	// final conversion
	var buf bytes.Buffer
	if err := md.Convert([]byte(out), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
