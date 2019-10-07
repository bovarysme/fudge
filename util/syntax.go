package util

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
)

func getStyle() (*chroma.Style, error) {
	style, err := chroma.NewStyle("ayu-dark", chroma.StyleEntries{
		chroma.Background:      " bg:#0a0e14",
		chroma.Text:            "#b3b1ad",
		chroma.Keyword:         "#ff8f40",
		chroma.KeywordConstant: "#ffee99",
		chroma.KeywordType:     "#39bae6",
		chroma.LiteralNumber:   "#ffee99",
		chroma.LiteralString:   "#c2d94c",
	})

	if err != nil {
		return nil, err
	}

	return style, nil
}

func Highlight(filename string, r io.ReadCloser) (string, error) {
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

	style, err := getStyle()
	if err != nil {
		return "", err
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
