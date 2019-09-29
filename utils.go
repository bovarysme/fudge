package main

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

func highlight(filename string, r io.ReadCloser) (string, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	contents := string(b)

	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Analyse(contents)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := styles.Get("github")
	if style == nil {
		style = styles.Fallback
	}

	iterator, err := lexer.Tokenise(nil, contents)
	if err != nil {
		return "", err
	}

	buffer := new(bytes.Buffer)
	formatter := html.New(html.TabWidth(4), html.WithLineNumbers())

	err = formatter.Format(buffer, style, iterator)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
