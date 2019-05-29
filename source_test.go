package xls2sheets

import (
	"strconv"
	"testing"
	"time"
)

func Test_generateName(t *testing.T) {
	type args struct {
		prefix string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"with prefix", args{"prefix"}, "prefix" + strconv.Itoa(int(time.Now().Unix())) + ".xlsx"},
		{"no prefix", args{""}, strconv.Itoa(int(time.Now().Unix())) + ".xlsx"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateName(tt.args.prefix); got != tt.want {
				t.Errorf("generateName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getFilename(t *testing.T) {
	type args struct {
		loc string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"just file", args{"file://sample.xls"}, "sample.xls", false},
		{"full path", args{"file:///tmp/subdir/subdir2/file.txt"}, "/tmp/subdir/subdir2/file.txt", false},
		{"current dir file", args{"file://./for_anna.xls"}, "for_anna.xls", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFilename(tt.args.loc)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFilename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
