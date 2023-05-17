package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomString(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "length 1",
			args: args{
				n: 1,
			},
			wantErr: false,
		},
		{
			name: "length 5",
			args: args{
				n: 5,
			},
			wantErr: false,
		},
		{
			name: "length 10",
			args: args{
				n: 10,
			},
			wantErr: false,
		},
		{
			name: "negative length",
			args: args{
				n: -1,
			},
			wantErr: true,
		},
		{
			name: "zero length",
			args: args{
				n: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateRandomString(tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRandomString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.args.n < 0 {
				assert.Len(t, got, 0)
			} else {
				assert.Len(t, got, tt.args.n)
			}
		})
	}
}
