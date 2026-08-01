package models

import (
	"errors"
	"fmt"

	"github.com/taiidani/no-time-to-explain/internal/bluesky"
)

func (q *Queries) ValidateFeed(f Feed) error {
	var ret error

	client := bluesky.NewBlueskyClient()
	user, err := client.GetUser(f.Author)
	if err != nil {
		ret = errors.Join(fmt.Errorf("could not look up user %q: %w", f.Author, err))
	} else {
		f.AuthorSourceID = user.DID
	}

	return ret
}

func (f *Feed) URL() string {
	return fmt.Sprintf("https://bsky.app/profile/%s", f.AuthorSourceID)
}
