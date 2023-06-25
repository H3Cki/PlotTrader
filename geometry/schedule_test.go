package geometry_test

import (
	"testing"
	"time"

	"github.com/H3Cki/Plotor/geometry"
	"github.com/stretchr/testify/assert"
)

func TestSchedule_InRange(t *testing.T) {
	type fields struct {
		Since time.Time
		Until time.Time
		Plot  geometry.Plot
	}

	tests := []struct {
		name   string
		fields fields
		t      time.Time
		want   bool
	}{
		{
			name:   "always valid - zero time",
			fields: fields{},
			t:      time.Time{},
			want:   true,
		},
		{
			name:   "always valid - max time",
			fields: fields{},
			t:      time.Unix(1<<63-1, 0),
			want:   true,
		},
		{
			name:   "since - exact",
			fields: fields{Since: time.Unix(10, 0)},
			t:      time.Unix(10, 0),
			want:   true,
		},
		{
			name:   "since - after",
			fields: fields{Since: time.Unix(10, 0)},
			t:      time.Unix(11, 0),
			want:   true,
		},
		{
			name:   "since - before",
			fields: fields{Since: time.Unix(10, 0)},
			t:      time.Unix(9, 0),
			want:   false,
		},
		{
			name:   "until - exact",
			fields: fields{Until: time.Unix(10, 0)},
			t:      time.Unix(10, 0),
			want:   false,
		},
		{
			name:   "until - after",
			fields: fields{Until: time.Unix(10, 0)},
			t:      time.Unix(11, 0),
			want:   false,
		},
		{
			name:   "until - before",
			fields: fields{Until: time.Unix(10, 0)},
			t:      time.Unix(9, 0),
			want:   true,
		},
		{
			name:   "since, until - exact since",
			fields: fields{Since: time.Unix(10, 0), Until: time.Unix(20, 0)},
			t:      time.Unix(10, 0),
			want:   true,
		},
		{
			name:   "since, until - exact until",
			fields: fields{Since: time.Unix(10, 0), Until: time.Unix(20, 0)},
			t:      time.Unix(20, 0),
			want:   false,
		},
		{
			name:   "since, until - in between",
			fields: fields{Since: time.Unix(10, 0), Until: time.Unix(20, 0)},
			t:      time.Unix(15, 0),
			want:   true,
		},
		{
			name:   "valid since - before",
			fields: fields{Since: time.Unix(10, 0), Until: time.Unix(20, 0)},
			t:      time.Unix(9, 0),
			want:   false,
		},
		{
			name:   "valid since - after",
			fields: fields{Since: time.Unix(10, 0), Until: time.Unix(20, 0)},
			t:      time.Unix(9, 0),
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &geometry.Schedule{
				Since: tt.fields.Since,
				Until: tt.fields.Until,
				Plot:  tt.fields.Plot,
			}
			got := v.InRange(tt.t)
			assert.Equal(t, tt.want, got)
		})
	}
}
