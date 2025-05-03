package internal

import (
	"reflect"
	"testing"
	"time"

	"github.com/taiidani/no-time-to-explain/internal/bluesky"
	"github.com/taiidani/no-time-to-explain/internal/models"
)

func Test_filterPosts(t *testing.T) {
	tm := time.Now()
	postSecondAgo := bluesky.FeedPost{
		Record: bluesky.FeedPostRecord{
			Text:      "second ago",
			CreatedAt: tm.Add(time.Second * -1),
		},
	}
	postMinuteAgo := bluesky.FeedPost{
		Record: bluesky.FeedPostRecord{
			Text:      "minute ago",
			CreatedAt: tm.Add(time.Minute * -1),
		},
	}
	postTwoMinuteAgo := bluesky.FeedPost{
		Record: bluesky.FeedPostRecord{
			Text:      "two minutes ago",
			CreatedAt: tm.Add(time.Minute * -2),
		},
	}
	postHourAgo := bluesky.FeedPost{
		Record: bluesky.FeedPostRecord{
			Text:      "hour ago",
			CreatedAt: tm.Add(time.Hour * -1),
		},
	}

	type args struct {
		feed  models.Feed
		posts []bluesky.FeedPostEntry
	}
	tests := []struct {
		name string
		args args
		want []bluesky.FeedPostEntry
	}{
		{
			name: "default",
			args: args{
				feed: models.Feed{LastMessage: tm.Add(time.Hour * -2)},
				posts: []bluesky.FeedPostEntry{
					{Post: postSecondAgo}, // This might not have its embeds processed yet
					{Post: postMinuteAgo},
					{Post: postTwoMinuteAgo},
					{Post: postHourAgo},
				},
			},
			want: []bluesky.FeedPostEntry{
				{Post: postMinuteAgo},
				{Post: postTwoMinuteAgo},
				{Post: postHourAgo},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterPosts(tt.args.feed, tt.args.posts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterPosts() = %v, want %v", got, tt.want)
			}
		})
	}
}
