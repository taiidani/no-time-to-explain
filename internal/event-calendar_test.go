package internal

import (
	"reflect"
	"testing"
	"time"
)

func Test_parseEventCalendarStartTime(t *testing.T) {
	type args struct {
		tm string
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "golden",
			args: args{tm: `<t:1716073200:F>\n<t:1716073200:R>`},
			want: time.Unix(1716073200, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseEventCalendarStartTime(tt.args.tm); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseEventCalendarStartTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
