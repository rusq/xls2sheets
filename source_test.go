package xls2sheets

import (
	"io/ioutil"
	"os"
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

func Test_filename(t *testing.T) {
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
		{"current dir file", args{"file://./some_old_file.xls"}, "some_old_file.xls", false},
		{"url parse error", args{"://what_proto_is_that?"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filename(tt.args.loc)
			if (err != nil) != tt.wantErr {
				t.Errorf("filename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("filename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fileType(t *testing.T) {
	f, err := ioutil.TempFile("", tempFilePrefix+"*")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer func() {
		os.Remove(f.Name())
	}()
	type args struct {
		loc string
	}
	tests := []struct {
		name string
		args args
		want srcType
	}{
		{"local file with schema", args{"file://tea_with_lemon.xls"}, srcFile},
		{"local file without schema", args{f.Name()}, srcFile},
		{"remote file", args{"https://shadywebsite.com/README.xlsx"}, srcWeb},
		{"google sheets id", args{"1jw2phhb11w6vKw5nkklWtSZHOxADIsXcyRgVLyea4ak"}, srcGSheet},
		{"unknown", args{""}, srcUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fileType(tt.args.loc); got != tt.want {
				t.Errorf("fileType() = %v, want %v", got, tt.want)
			}
		})
	}
}
