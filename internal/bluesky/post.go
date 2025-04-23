package bluesky

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type FeedData struct {
	Feed []FeedPostEntry `json:"feed"`
}

type FeedPostEntry struct {
	Post FeedPost `json:"post"`
}

// FeedPost contains the content of a single post in an author's feed
// Example URL for a post: https://bsky.app/profile/destinythegame.bungie.net/post/3lnape7mfxs27
type FeedPost struct {
	Author      FeedPostAuthor `json:"author"`
	CID         string         `json:"cid"`
	Embed       any            `json:"embed"`
	IndexedAt   time.Time      `json:"indexedAt"`
	Labels      []any          `json:"labels"`
	LikeCount   int            `json:"likeCount"`
	QuoteCount  int            `json:"quoteCount"`
	Record      FeedPostRecord `json:"record"`
	ReplyCount  int            `json:"replyCount"`
	RepostCount int            `json:"repostCount"`
	URI         string         `json:"uri"`
}

// FeedPostAuthor contains the bulk of the information about the author of a given `FeedPost`.
type FeedPostAuthor struct {
	Avatar      string    `json:"avatar"`
	CreatedAt   time.Time `json:"createdAt"`
	DID         string    `json:"did"`
	DisplayName string    `json:"displayName"`
	Handle      string    `json:"handle"`
	Labels      []any     `json:"labels"`
}

// FeedPostRecord contains the bulk of the information about the contents of a given `FeedPost`.
type FeedPostRecord struct {
	Type      string    `json:"$type"`
	CreatedAt time.Time `json:"createdAt"`
	Embed     any       `json:"embed"`
	Facets    any       `json:"facets"`
	Text      string    `json:"text"`
}

// GetUserFeed fetches user feed data by handle
func (c *BlueskyClient) GetUserFeed(handle string) (*FeedData, error) {
	params := url.Values{}
	params.Add("actor", handle)
	url := fmt.Sprintf("%s/xrpc/app.bsky.feed.getAuthorFeed?%s", c.BaseURL, params.Encode())

	resp, err := c.HttpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(bodyBytes))
	}

	var ret FeedData
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}

	return &ret, nil
}

func (p *FeedPost) URL() string {
	atParts := strings.Split(p.URI, "/")
	postID := atParts[len(atParts)-1]

	// Example: https://bsky.app/profile/destinythegame.bungie.net/post/3lnape7mfxs27
	return fmt.Sprintf("https://bsky.app/profile/%s/post/%s", p.Author.Handle, postID)
}
