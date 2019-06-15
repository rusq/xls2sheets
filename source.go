package xls2sheets

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

const (
	// MIME types
	// Google sheets
	gsheetMIME = "application/vnd.google-apps.spreadsheet"
	// Microsoft Excel .xlsx
	xlsxMIME = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	// Microsoft Excel .xls
	xlsMIME = "application/vnd.ms-excel"

	// tempFilePrefix used for temporary file uploads
	tempFilePrefix = "xls2sheets$"
)

var extensionMIMEmap = map[string]string{
	".xlsx": xlsxMIME,
	".xls":  xlsMIME,
}

// SourceFile contains the information about the source file and
// address + range of cells to copy
type SourceFile struct {
	// Location specifies the file location
	// Valid values:
	//
	// 		https://www.example.com/dataset.xlsx
	//		file://MyWorkbook.xlsx
	FileLocation string `yaml:"location"`
	// SheetAddress is the address within the source workbook.
	// I.e. "Data!A1:U"
	SheetAddressRange []string `yaml:"address_range"`

	file *drive.File // handle to uploaded file
}

var (
	errNothingToDelete = errors.New("delete called before upload")
)

// fetchFromWeb loads a source file on a remote server
func fetchFromWeb(uri string) (io.ReadCloser, error) {
	tlsConfig := tls.Config{
		InsecureSkipVerify: true,
	}
	transport := &http.Transport{TLSClientConfig: &tlsConfig}

	insecureClient := &http.Client{Transport: transport}
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	resp, err := insecureClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func fetch(loc string) (io.ReadCloser, error) {
	switch {
	case strings.HasPrefix(strings.ToLower(loc), "file://"):
		filename, err := getFilename(loc)
		if err != nil {
			return nil, err
		}
		return os.Open(filename)
	case strings.Contains(loc, "://"):
		return fetchFromWeb(loc)
	default:
		// no schema, defaults to local file
		return os.Open(loc)
	}
	// UNREACHABLE
}

func getFilename(loc string) (string, error) {
	url, err := url.Parse(loc)
	if err != nil {
		return "", err
	}
	// dirty hax
	return filepath.Join(url.Host, url.Path), nil
}

// FetchAndUpload downloads the file from source and uploads it to Google
// Drive
func (sf *SourceFile) FetchAndUpload(client *http.Client) (string, error) {
	log.Printf("+ trying to open: %s", sf.FileLocation)
	src, err := fetch(sf.FileLocation)
	if err != nil {
		return "", err
	}
	defer src.Close()
	return sf.upload(client, src)
}

// Delete deletes the temporary file from the google drive.
func (sf *SourceFile) Delete(client *http.Client) error {
	// if the file handle is nil, then upload function hasn't been called yet
	if sf.file == nil {
		return errNothingToDelete
	}
	srv, err := drive.New(client)
	if err != nil {
		return err
	}
	if err := srv.Files.Delete(sf.file.Id).Do(); err != nil {
		return err
	}
	// clearing the file ID so that consequent calls would now that the file
	// does not exist
	sf.file = nil
	return nil
}

// upload uploads the source data to temporary google spreadsheet on
// google drive, so that it would be possible to copy data from it.
func (sf *SourceFile) upload(client *http.Client, sourceData io.Reader) (string, error) {
	srv, err := drive.New(client)
	if err != nil {
		return "", err
	}
	// target file name and MIME type format, so that Google Drive would
	// convert the source excel file to Google Sheets format
	file := drive.File{
		Name:     generateName(tempFilePrefix, filepath.Ext(sf.FileLocation)),
		MimeType: gsheetMIME,
	}
	// content type is necessary for google drive to convert the file to
	sf.file, err = srv.Files.
		Create(&file).
		Media(
			sourceData, // source file data
			googleapi.ContentType(getMIME(sf.FileLocation)), // source file MIME type
		).
		Do()
	if err != nil {
		return "", err
	}

	return sf.file.Id, err
}

// generateName generates a temporary filename.
func generateName(prefix string, extension string) string {
	epoch := time.Now().Unix()
	return fmt.Sprintf("%s%d%s", prefix, epoch, extension)
}

// getMIME returns the MIME for the given filename.
func getMIME(filename string) string {
	mime, ok := extensionMIMEmap[strings.ToLower(filepath.Ext(filename))]
	if !ok {
		// BUG: defaults to xlsx mime, maybe will need to reconsider
		return xlsxMIME
	}
	return mime
}
