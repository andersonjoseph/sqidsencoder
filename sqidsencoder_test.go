package sqidsencoder

import (
	"reflect"
	"testing"

	"github.com/sqids/sqids-go"
)

func TestEncode(t *testing.T) {
	type user struct {
		ID       int64  `json:"id" sqids:"encode"`
		Name     string `json:"name"`
		Username string `json:"username"`
	}

	type encodedUser struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
	}

	type args struct {
		v any
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
			name: "encode numeric id",
			args: args{
				v: user{
					ID:       1,
					Name:     "Anderson",
					Username: "andersonjoseph",
				},
			},
			want: encodedUser{
				ID:       encodeIDHelper(1, s, t),
				Name:     "Anderson",
				Username: "andersonjoseph",
			},
		},
	}

	encoder := New(s)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := encodedUser{}
			err := encoder.Encode(tt.args.v, &res)

			if tt.wantErr != (err != nil) {
				t.Errorf("Encode error: %s = %v, want %v", tt.name, err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.want, res) {
				t.Errorf("Test failed: %s = %v, want %v", tt.name, res, tt.want)
			}
		})
	}
}


func encodeIDHelper(id int, s sqidsInterface, t *testing.T) string {
	t.Helper()

	r, e := s.Encode([]uint64{uint64(id)})

	if e != nil {
		t.Fatal(e)
	}

	return r
}
