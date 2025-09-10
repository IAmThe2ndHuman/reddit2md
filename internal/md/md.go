package md

import (
	"os"
	"reddit2md/internal/api"

	"github.com/nao1215/markdown"
)

func WritePostToMarkdown(p *api.LinkedPost, title, path string) error {
	f, err := os.Create(path)

	if err != nil {
		return err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	return markdown.NewMarkdown(f).
		H1(title).
		H2(p.Title).
		PlainTextf("%s %s", markdown.Bold("Post author:"), p.Author).
		LF().
		PlainText(p.Content).
		Build()
}
