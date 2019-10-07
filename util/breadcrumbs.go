package util

import (
	"fmt"
	"strings"
)

type Breadcrumb struct {
	Text string
	Link string
}

func Breadcrumbs(name, path string) []*Breadcrumb {
	crumbs := []*Breadcrumb{
		&Breadcrumb{
			Text: name,
			Link: fmt.Sprintf("/%s", name),
		},
	}

	current := fmt.Sprintf("/%s/tree", name)
	parts := strings.Split(path, "/")

	for _, part := range parts {
		if part == "" {
			continue
		}

		current = fmt.Sprintf("%s/%s", current, part)

		crumb := &Breadcrumb{
			Text: part,
			Link: current,
		}
		crumbs = append(crumbs, crumb)
	}

	crumbs[len(crumbs)-1].Link = ""

	return crumbs
}
