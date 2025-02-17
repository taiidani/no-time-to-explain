package bot

import (
	"reflect"
	"testing"
	"time"
)

func Test_parseTimestamp(t *testing.T) {
	pacific, err := parseTimezone("PST")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		opts interactionState
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "golden",
			args: args{
				opts: interactionState{
					Date: "2024-01-02",
					Time: "01:02:03 AM",
					TZ:   "UTC",
				},
			},
			want: time.Date(2024, time.January, 2, 1, 2, 3, 0, time.UTC),
		},
		{
			name: "no-hyphens",
			args: args{
				opts: interactionState{
					Date: "20251010",
					Time: "13:26:45",
					TZ:   "UTC",
				},
			},
			wantErr: true,
		},
		{
			name: "no-seconds",
			args: args{
				opts: interactionState{
					Date: "20251010",
					Time: "13:26",
					TZ:   "UTC",
				},
			},
			wantErr: true,
		},
		{
			name: "short-hour",
			args: args{
				opts: interactionState{
					Date: "2025-10-10",
					Time: "3:26:00 AM",
					TZ:   "UTC",
				},
			},
			want: time.Date(2025, time.October, 10, 3, 26, 0, 0, time.UTC),
		},
		{
			name: "late-hour",
			args: args{
				opts: interactionState{
					Date: "2025-10-10",
					Time: "11:59:00 PM",
					TZ:   "UTC",
				},
			},
			want: time.Date(2025, time.October, 10, 23, 59, 0, 0, time.UTC),
		},
		{
			name: "pacific-tz",
			args: args{
				opts: interactionState{
					Date: "2025-10-10",
					Time: "11:59:00 PM",
					TZ:   "PST",
				},
			},
			want: time.Date(2025, time.October, 10, 23, 59, 0, 0, pacific),
		},
		{
			name: "invalid-date",
			args: args{
				opts: interactionState{
					Date: "202510101",
					Time: "23:59",
					TZ:   "UTC",
				},
			},
			want:    time.Time{},
			wantErr: true,
		},
		{
			name: "invalid-time",
			args: args{
				opts: interactionState{
					Date: "20251010",
					Time: "23:59:",
					TZ:   "UTC",
				},
			},
			want:    time.Time{},
			wantErr: true,
		},
		{
			name: "empty-date",
			args: args{
				opts: interactionState{
					Date: "",
					Time: "23:59",
					TZ:   "UTC",
				},
			},
			want:    time.Time{},
			wantErr: true,
		},
		{
			name: "empty-time",
			args: args{
				opts: interactionState{
					Date: "20251010",
					Time: "",
					TZ:   "UTC",
				},
			},
			want:    time.Time{},
			wantErr: true,
		},
		{
			name: "invalid-tz",
			args: args{
				opts: interactionState{
					Date: "20251010",
					Time: "23:59",
					TZ:   "Nowhere",
				},
			},
			want:    time.Time{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimestamp(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}
