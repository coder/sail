package main

import "testing"

func Test_resolvePath(t *testing.T) {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolvePath(tt.args.homedir, tt.args.path); got != tt.want {
				t.Errorf("resolvePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
