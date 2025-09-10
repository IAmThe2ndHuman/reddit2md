package api

import (
	"context"
	"iter"
	"strings"
	"time"
)

// LinkedPost represents a Reddit post that may contain a link to the next post in a series.
type LinkedPost struct {
	Title   string
	Content string
	URL     string
	Author  string
	Created time.Time
	NextURL string
}

// SanitizeContent cleans the post content to make it visually nicer.
//
// Sanitation includes the following steps:
// - Remove zero-width spaces I've seen in some posts
// - Remove lines that contain Next links
// - Remove leading and trailing whitespace
func (re *LinkedPost) SanitizeContent() {
	var keptLines []string

	for _, line := range strings.Split(re.Content, "\n") {
		if nextRegex.MatchString(line) {
			continue
		}

		keptLines = append(keptLines, line)
	}

	re.Content = strings.Join(keptLines, "\n")

	re.Content = strings.ReplaceAll(re.Content, "&amp;#x200B;\n", "")
	re.Content = strings.ReplaceAll(re.Content, "&amp;#x200B;", "")

	re.Content = strings.TrimSpace(re.Content)
}

// Iter returns an iterator that walks through the linked posts starting from the current one. To prevent
// rate limiting, the caller should delay between requests.
func (re *LinkedPost) Iter(ctx context.Context, client *RedditClient) iter.Seq2[*LinkedPost, error] {
	return func(yield func(*LinkedPost, error) bool) {
		current := re

		for {
			if !yield(current, nil) {
				return
			}

			if current.NextURL == "" {
				return
			}

			next, err := client.GetLinkedPost(ctx, current.NextURL)

			if err != nil {
				yield(nil, err)
				return
			}

			current = next
		}
	}
}
