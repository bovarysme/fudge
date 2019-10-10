package util

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
)

func getFormatter() *html.Formatter {
	return html.New(html.TabWidth(4), html.WithClasses(),
		html.WithLineNumbers(), html.LineNumbersInTable())
}

func getStyle() (*chroma.Style, error) {
	style, err := chroma.NewStyle("ayu-dark", chroma.StyleEntries{
		chroma.Background:         " bg:#0a0e14",
		chroma.Text:               "#b3b1ad",
		chroma.Comment:            "#626a73",
		chroma.Error:              "#ff3333",
		chroma.GenericDeleted:     "#d96c75",
		chroma.GenericInserted:    "#91b362",
		chroma.Keyword:            "#ff8f40",
		chroma.KeywordConstant:    "#ffee99",
		chroma.KeywordType:        "#39bae6",
		chroma.LiteralNumber:      "#ffee99",
		chroma.LiteralString:      "#c2d94c",
		chroma.LiteralStringChar:  "#95e6cb",
		chroma.LiteralStringOther: "#95e6cb",
		chroma.LiteralStringRegex: "#95e6cb",
		chroma.NameAttribute:      "#ffb454",
		chroma.NameClass:          "#ffb454",
		chroma.NameDecorator:      "#e6b673",
		chroma.NameNamespace:      "#ffb454",
		chroma.NameTag:            "#39bae6",
		chroma.OperatorWord:       "#f29668",
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
	formatter := getFormatter()

	err = formatter.Format(buffer, style, iterator)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func WriteCSS(w io.Writer) error {
	style, err := getStyle()
	if err != nil {
		return err
	}

	formatter := getFormatter()
	err = formatter.WriteCSS(w, style)

	return err
}
