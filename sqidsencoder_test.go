package sqidsencoder

import (
	"reflect"
	"testing"

	"github.com/sqids/sqids-go"
)

func TestEncode(t *testing.T) {
	type args struct {
		src any
		dst any
	}

	s, err := sqids.New()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "encoding numeric ID",
			args: args{
				src: struct {
					ID       uint64 `json:"id" sqids:"encode"`
					Username string `json:"username"`
				}{
					ID:       1,
					Username: "andersonjoseph",
				},
				dst: &struct {
					ID       string
					Username string
				}{},
			},
			want: &struct {
				ID       string
				Username string
			}{
				ID:       encodeIDHelper(t, s, 1),
				Username: "andersonjoseph",
			},
		},
		{
			name: "encoding numeric ID without a sqids tag returns an error",
			args: args{
				src: struct {
					ID       uint64
					Username string
				}{},
				dst: &struct {
					ID       string
					Username string
				}{},
			},
			want: &struct {
				ID       string
				Username string
			}{},
			wantErr: true,
		},
	}

	encoder := New(s)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := encoder.Encode(tt.args.src, tt.args.dst)

			if err != nil {
				t.Log(err)
			}

			if tt.wantErr != (err != nil) {
				t.Errorf("Encode error: %s = %v, want %v", tt.name, err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.want, tt.args.dst) {
				t.Errorf("Test failed: %s = %v, want %v", tt.name, tt.args.dst, tt.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	type args struct {
		src any
		dst any
	}

	s, err := sqids.New()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "decode numeric id of the property ID",
			args: args{
				src: struct {
					ID       string `sqids:"decode"`
					Username string `json:"username"`
				}{
					ID:       encodeIDHelper(t, s, 1),
					Username: "andersonjoseph",
				},
				dst: &struct {
					ID       uint64
					Username string
				}{},
			},
			want: &struct {
				ID       uint64
				Username string
			}{
				ID:       1,
				Username: "andersonjoseph",
			},
			wantErr: false,
		},
	}

	decoder := New(s)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decoder.Decode(tt.args.src, tt.args.dst)

			if err != nil {
				t.Log(err)
			}

			if tt.wantErr != (err != nil) {
				t.Errorf("Encode error: %s = %v, want %v", tt.name, err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.want, tt.args.dst) {
				t.Errorf("Test failed: %s = %v, want %v", tt.name, tt.args.dst, tt.want)
			}
		})
	}
}

func encodeIDHelper(t *testing.T, s sqidsInterface, id int) string {
	t.Helper()

	r, e := s.Encode([]uint64{uint64(id)})

	if e != nil {
		t.Fatal(e)
	}

	return r
}
