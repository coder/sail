package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_resolvePath(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	type args struct {
		homedir string
		path    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Abs", args{"/home/ammar", "/home/ammar/test"}, "/home/ammar/test"},
		{"NoRel", args{"/home/ammar", wd}, wd},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolvePath(tt.args.homedir, tt.args.path); got != tt.want {
				t.Errorf("resolvePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
