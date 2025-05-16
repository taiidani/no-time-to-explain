package bot

import (
	"testing"

	"github.com/taiidani/no-time-to-explain/internal/models"
)

func Test_responseForTrigger(t *testing.T) {
	fixtures := []models.Message{
		{Trigger: "[jJ]esus", Response: "You mean Bees-us?"},
		{Trigger: "multi", Response: "Response Foo"},
		{Trigger: "[mM]ulti", Response: "Response Bar"},
		{Trigger: "[mM]ulti", Response: "Response Baz"},
		{Trigger: "^ping$", Response: "pong"},
	}

	type args struct {
		input string
	}
	tests := []struct {
		name     string
		messages []models.Message
		seed     int64
		args     args
		want     string
	}{
		{
			name:     "jesus",
			messages: fixtures,
			args: args{
				input: "jesus",
			},
			want: "You mean Bees-us?",
		},
		{
			name:     "jesus christ",
			messages: fixtures,
			args: args{
				input: "Jesus christ",
			},
			want: "You mean Bees-us?",
		},
		{
			name:     "embedded jesus",
			messages: fixtures,
			args: args{
				input: "Holy jesus christ",
			},
			want: "You mean Bees-us?",
		},
		{
			name:     "ping",
			messages: fixtures,
			args: args{
				input: "ping",
			},
			want: "pong",
		},
		{
			name:     "Ping",
			messages: fixtures,
			args: args{
				input: "Ping",
			},
			want: "",
		},
		{
			name:     "embedded ping",
			messages: fixtures,
			args: args{
				input: "A ping inside.",
			},
			want: "",
		},
		{
			name:     "empty db",
			messages: nil,
			args: args{
				input: "A ping inside.",
			},
			want: "",
		},
		{
			name:     "multi-1",
			messages: fixtures,
			seed:     5,
			args: args{
				input: "A trigger for multiple responses.",
			},
			want: "Response Foo",
		},
		{
			name:     "multi-2",
			messages: fixtures,
			seed:     2,
			args: args{
				input: "A trigger for multiple responses.",
			},
			want: "Response Bar",
		},
		{
			name:     "multi-3",
			messages: fixtures,
			seed:     1,
			args: args{
				input: "A trigger for multiple responses.",
			},
			want: "Response Baz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Commands{}

			if tt.seed > 0 {
				responseSeeder.Seed(tt.seed)
			}

			if got := c.responseForTrigger(tt.messages, tt.args.input); got != tt.want {
				t.Errorf("responseForTrigger() = %v, want %v", got, tt.want)
			}
		})
	}
}
