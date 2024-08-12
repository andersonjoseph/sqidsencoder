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
		{
			name: "passing a non pointer struct as dst returns an error",
			args: args{
				src: struct {
					ID       uint64 `sqids:"encode"`
					Username string
				}{},
				dst: struct {
					ID       string
					Username string
				}{},
			},
			want: struct {
				ID       string
				Username string
			}{},
			wantErr: true,
		},
		{
			name: "passing a non struct as src returns an error",
			args: args{
				src: 123,
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
		{
			name: "passing a dst struct with a encoded field as non string",
			args: args{
				src: struct {
					ID       uint64 `sqids:"encode"`
					Username string
				}{},
				dst: &struct {
					ID       uint64
					Username string
				}{},
			},
			want: &struct {
				ID       uint64
				Username string
			}{},
			wantErr: true,
		},
		{
			name: "passing a dst without the decoded property returns an error",
			args: args{

				src: struct {
					ID       uint64 `sqids:"encode"`
					Username string
				}{},
				dst: &struct {
					Username string
				}{},
			},
			want: &struct {
				Username string
			}{},
			wantErr: true,
		},
		{
			name: "encoding ID in nested structs",
			args: args{
				src: struct {
					ID   uint64 `sqids:"encode"`
					Item struct {
						ID   uint64 `sqids:"encode"`
						Name string
					} `sqids:"encode"`
				}{
					ID: 1,
					Item: struct {
						ID   uint64 `sqids:"encode"`
						Name string
					}{
						ID:   1,
						Name: "cool item",
					},
				},
				dst: &struct {
					ID   string
					Item struct {
						ID   string
						Name string
					}
				}{},
			},
			want: &struct {
				ID   string
				Item struct {
					ID   string
					Name string
				}
			}{
				ID: encodeIDHelper(t, s, 1),
				Item: struct {
					ID   string
					Name string
				}{
					ID:   encodeIDHelper(t, s, 1),
					Name: "cool item",
				},
			},
			wantErr: false,
		},
		{
			name: "ID to encode not int",
			args: args{
				src: struct {
					ID       string `sqids:"encode"`
					Username string
				}{
					ID:       "1",
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
			}{},
			wantErr: true,
		},
		{
			name: "encoding slices",
			args: args{
				src: struct {
					ID []uint64 `sqids:"encode"`
				}{
					ID: []uint64{1, 2, 3},
				},
				dst: &struct {
					ID []string
				}{},
			},
			want: &struct {
				ID []string
			}{
				ID: encodeIDsHelper(t, s, []uint64{1, 2, 3}),
			},
			wantErr: false,
		},
		{
			name: "encoding slice of structs",
			args: args{
				src: struct {
					Items []struct {
						ID uint64 `sqids:"encode"`
					} `sqids:"encode"`
				}{
					Items: []struct {
						ID uint64 `sqids:"encode"`
					}{
						{ID: 1},
						{ID: 2},
						{ID: 3},
					},
				},
				dst: &struct {
					Items []struct{ ID string }
				}{},
			},
			want: &struct {
				Items []struct{ ID string }
			}{
				Items: []struct{ ID string }{
					{ID: encodeIDHelper(t, s, 1)},
					{ID: encodeIDHelper(t, s, 2)},
					{ID: encodeIDHelper(t, s, 3)},
				},
			},
			wantErr: false,
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
			name: "decoding numeric ID",
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
		{
			name: "decoding numeric ID without a sqids tag returns an error",
			args: args{
				src: struct {
					ID       string
					Username string
				}{},
				dst: &struct {
					ID       uint64
					Username string
				}{},
			},
			want: &struct {
				ID       uint64
				Username string
			}{},
			wantErr: true,
		},
		{
			name: "passing a non pointer struct as dst returns an error",
			args: args{
				src: struct {
					ID       string `sqids:"decode"`
					Username string
				}{},
				dst: struct {
					ID       uint64
					Username string
				}{},
			},
			want: struct {
				ID       uint64
				Username string
			}{},
			wantErr: true,
		},
		{
			name: "passing a non struct as src returns an error",
			args: args{
				src: 123,
				dst: &struct {
					ID       uint64
					Username string
				}{},
			},
			want: &struct {
				ID       uint64
				Username string
			}{},
			wantErr: true,
		},
		{
			name: "passing a dst struct with a encoded field as non uint64",
			args: args{
				src: struct {
					ID       string `sqids:"decode"`
					Username string
				}{
					ID: encodeIDHelper(t, s, 1),
				},
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
		{
			name: "decoding ID in nested structs",
			args: args{
				src: struct {
					ID   string `sqids:"decode"`
					Item struct {
						ID   string `sqids:"decode"`
						Name string
					} `sqids:"decode"`
				}{
					ID: encodeIDHelper(t, s, 1),
					Item: struct {
						ID   string `sqids:"decode"`
						Name string
					}{
						ID:   encodeIDHelper(t, s, 1),
						Name: "cool item",
					},
				},
				dst: &struct {
					ID   uint64
					Item struct {
						ID   uint64
						Name string
					}
				}{},
			},
			want: &struct {
				ID   uint64
				Item struct {
					ID   uint64
					Name string
				}
			}{
				ID: 1,
				Item: struct {
					ID   uint64
					Name string
				}{
					ID:   1,
					Name: "cool item",
				},
			},
			wantErr: false,
		},
		{
			name: "ID to decode not string",
			args: args{
				src: struct {
					ID uint64 `sqids:"decode"`
				}{
					ID: 1,
				},
				dst: &struct {
					ID string
				}{},
			},
			want: &struct {
				ID string
			}{},
			wantErr: true,
		},
		{
			name: "decoding slices",
			args: args{
				src: struct {
					IDs []string `sqids:"decode"`
				}{
					IDs: encodeIDsHelper(t, s, []uint64{1, 2, 3}),
				},
				dst: &struct {
					IDs []uint64
				}{},
			},
			want: &struct {
				IDs []uint64
			}{
				IDs: []uint64{1, 2, 3},
			},
			wantErr: false,
		},
		{
			name: "decoding slice of structs",
			args: args{
				src: struct {
					Items []struct {
						ID string `sqids:"decode"`
					} `sqids:"decode"`
				}{
					Items: []struct {
						ID string `sqids:"decode"`
					}{
						{ID: encodeIDHelper(t, s, 1)},
						{ID: encodeIDHelper(t, s, 2)},
						{ID: encodeIDHelper(t, s, 3)},
					},
				},
				dst: &struct {
					Items []struct{ ID uint64 }
				}{},
			},
			want: &struct {
				Items []struct{ ID uint64 }
			}{
				Items: []struct{ ID uint64 }{
					{ID: 1},
					{ID: 2},
					{ID: 3},
				},
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

func encodeIDsHelper(t *testing.T, s sqidsInterface, ids []uint64) []string {
	t.Helper()

	out := make([]string, len(ids))

	for i := range ids {
		out[i] = encodeIDHelper(t, s, ids[i])
	}

	return out
}

func encodeIDHelper(t *testing.T, s sqidsInterface, id uint64) string {
	t.Helper()

	r, e := s.Encode([]uint64{id})

	if e != nil {
		t.Fatal(e)
	}

	return r
}
