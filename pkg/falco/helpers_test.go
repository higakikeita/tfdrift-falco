package falco

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStringField(t *testing.T) {
	tests := []struct {
		name   string
		fields map[string]string
		key    string
		want   string
	}{
		{
			name: "Direct Match",
			fields: map[string]string{
				"ct.user.name": "john",
			},
			key:  "ct.user.name",
			want: "john",
		},
		{
			name: "Case Insensitive Match",
			fields: map[string]string{
				"CT.User.Name": "jane",
			},
			key:  "ct.user.name",
			want: "jane",
		},
		{
			name: "Not Found",
			fields: map[string]string{
				"other.field": "value",
			},
			key:  "ct.user.name",
			want: "",
		},
		{
			name:   "Empty Map",
			fields: map[string]string{},
			key:    "ct.user.name",
			want:   "",
		},
		{
			name: "Multiple Fields with Case Variations",
			fields: map[string]string{
				"field1":      "value1",
				"Field2":      "value2",
				"ct.user.arn": "arn:aws:iam::123:user/test",
			},
			key:  "CT.USER.ARN",
			want: "arn:aws:iam::123:user/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStringField(tt.fields, tt.key)
			assert.Equal(t, tt.want, got)
		})
	}
}
