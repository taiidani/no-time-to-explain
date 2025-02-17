package bot

import (
	"testing"
	"time"

	"github.com/taiidani/no-time-to-explain/internal/data"
	"github.com/taiidani/no-time-to-explain/internal/models"
)

func Test_responseForTrigger(t *testing.T) {
	db := models.Messages{
		Messages: []models.Message{
			{Trigger: "[jJ]esus", Response: "You mean Bees-us?"},
			{Trigger: "^ping$", Response: "pong"},
		},
	}

	type args struct {
		input string
	}
	tests := []struct {
		name string
		db   *models.Messages
		args args
		want string
	}{
		{
			name: "jesus",
			db:   &db,
			args: args{
				input: "jesus",
			},
			want: "You mean Bees-us?",
		},
		{
			name: "jesus christ",
			db:   &db,
			args: args{
				input: "Jesus christ",
			},
			want: "You mean Bees-us?",
		},
		{
			name: "embedded jesus",
			db:   &db,
			args: args{
				input: "Holy jesus christ",
			},
			want: "You mean Bees-us?",
		},
		{
			name: "ping",
			db:   &db,
			args: args{
				input: "ping",
			},
			want: "pong",
		},
		{
			name: "Ping",
			db:   &db,
			args: args{
				input: "Ping",
			},
			want: "",
		},
		{
			name: "embedded ping",
			db:   &db,
			args: args{
				input: "A ping inside.",
			},
			want: "",
		},
		{
			name: "empty db",
			db:   nil,
			args: args{
				input: "A ping inside.",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Commands{
				db: &data.MemoryStore{Data: map[string][]byte{}},
			}
			c.db.Set(t.Context(), models.MessagesDBKey, tt.db, time.Hour)

			if got := c.responseForTrigger(t.Context(), tt.args.input); got != tt.want {
				t.Errorf("responseForTrigger() = %v, want %v", got, tt.want)
			}
		})
	}
}
