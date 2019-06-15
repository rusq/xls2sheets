package xls2sheets

import (
	"strconv"
	"testing"
	"time"
)

func Test_generateName(t *testing.T) {
	type args struct {
		prefix    string
		extension string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"with prefix", args{"prefix", ".xlsx"}, "prefix" + strconv.Itoa(int(time.Now().Unix())) + ".xlsx"},
		{"no prefix", args{"", ""}, strconv.Itoa(int(time.Now().Unix())) + ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateName(tt.args.prefix, tt.args.extension); got != tt.want {
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

func Test_getMIME(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"excel", args{"My Documents/filename.xls"}, xlsMIME},
		{"xlsx", args{"My Documents/filename.xlsx"}, xlsxMIME},
		{"unknown", args{"/images/facepalm.jpg"}, xlsxMIME},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMIME(tt.args.filename); got != tt.want {
				t.Errorf("getMIME() = %v, want %v", got, tt.want)
			}
		})
	}
}
