package api

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/sethjones/go-reddit/v2/reddit"
)

// Example: [Next](https://www.reddit.com/r/SomeSub/comments/abcdef/some_post/)
var nextRegex = regexp.MustCompile(`\[(?i:Next)]\((.*?reddit\.com.*?)\)`)

// RedditClient wraps the reddit.Client from the go-reddit library.
type RedditClient struct {
	client *reddit.Client
}

// NewClient creates a new RedditClient with the provided credentials.
func NewClient(id, secret string) (*RedditClient, error) {
	credentials := reddit.Credentials{
		ID:     id,
		Secret: secret,
	}

	client, err := reddit.NewClient(credentials, reddit.WithApplicationOnlyOAuth(true))

	if err != nil {
		return nil, fmt.Errorf("failed to create Reddit client: %w", err)
	}

	return &RedditClient{client: client}, nil
}

// GetLinkedPost retrieves a LinkedPost from Reddit given its permalink.
//
// A post is considered a "linked post" if it contains a link labeled "Next" in its content,
// which points to the next post in a series.
// TODO: add option to look for continuation in comments (some posts continue in comments)
// TODO: possibly keep a list of some popular comments that add to the story
func (rc *RedditClient) GetLinkedPost(ctx context.Context, permalink string) (*LinkedPost, error) {
	id := postIDFromPermalink(permalink)

	post, _, err := rc.client.Post.Get(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	nextUrl := extractNextURL(post.Post.Body)

	linkedPost := &LinkedPost{
		Title:   post.Post.Title,
		Content: post.Post.Body,
		URL:     permalink,
		Author:  post.Post.Author,
		Created: post.Post.Created.Time,
		NextURL: nextUrl,
	}

	linkedPost.SanitizeContent()

	return linkedPost, nil
}

// postIDFromPermalink extracts the post ID from a Reddit permalink.
func postIDFromPermalink(permalink string) string {
	split := strings.Split(permalink, "/")

	return split[len(split)-3]
}

// extractNextURL searches the content for a link labeled "Next" and returns the URL if found.
func extractNextURL(content string) string {
	matches := nextRegex.FindStringSubmatch(content)

	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}
