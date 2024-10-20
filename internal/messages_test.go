package internal

import "testing"

func Test_responseForTrigger(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "jesus",
			args: args{
				input: "jesus",
			},
			want: "You mean Bees-us?",
		},
		{
			name: "jesus christ",
			args: args{
				input: "jesus christ",
			},
			want: "You mean Bees-us?",
		},
		{
			name: "embedded jesus",
			args: args{
				input: "Holy jesus christ",
			},
			want: "You mean Bees-us?",
		},
		{
			name: "ping",
			args: args{
				input: "ping",
			},
			want: "pong",
		},
		{
			name: "Ping",
			args: args{
				input: "Ping",
			},
			want: "",
		},
		{
			name: "embedded ping",
			args: args{
				input: "A ping inside.",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := responseForTrigger(tt.args.input); got != tt.want {
				t.Errorf("responseForTrigger() = %v, want %v", got, tt.want)
			}
		})
	}
}
