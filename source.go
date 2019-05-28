package xls2sheets

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

const (
	// MIME types
	gsheetMIME = "application/vnd.google-apps.spreadsheet"
	xlsxMIME   = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

	// tempFilePrefix used for temporary file uploads
	tempFilePrefix = "xls2sheets$"
)

// SourceFile contains the information about the source file and
// address + range of cells to copy
type SourceFile struct {
	// Location specifies the file location
	// Valid values:
	//
	// 		https://www.example.com/dataset.xlsx
	//		file://MyWorkbook.xlsx  -- not implemented yet!
	FileLocation string `yaml:"location"`
	// SheetAddress is the address within the source workbook.
	// I.e. "Data!A1:U"
	SheetAddressRange []string `yaml:"address_range"`

	file *drive.File // handle to uploaded file
}

var (
	errNothingToDelete = errors.New("delete called before upload")
)

// fetch opens a source file
func (sf *SourceFile) fetch() (io.ReadCloser, error) {
	tlsConfig := tls.Config{
		InsecureSkipVerify: true,
	}
	transport := &http.Transport{TLSClientConfig: &tlsConfig}

	insecureClient := &http.Client{Transport: transport}
	req, err := http.NewRequest(http.MethodGet, sf.FileLocation, nil)
	if err != nil {
		return nil, err
	}
	resp, err := insecureClient.Do(req)
	// resp, err := http.Get(sf.FileLocation)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// FetchAndUpload downloads the file from source and uploads it to Google
// Drive
func (sf *SourceFile) FetchAndUpload(client *http.Client) (string, error) {
	src, err := sf.fetch()
	if err != nil {
		return "", err
	}
	defer src.Close()
	return sf.upload(client, src)
}

// Delete deletes the temporary file from the google drive
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
// google drive.
func (sf *SourceFile) upload(client *http.Client, sourceData io.Reader) (string, error) {
	srv, err := drive.New(client)
	if err != nil {
		return "", err
	}
	// target file name and MIME type format, so that Google Drive would
	// convert the source excel file to Google Sheets format
	file := drive.File{
		Name:     generateName(tempFilePrefix),
		MimeType: gsheetMIME,
	}
	// content type is necessary for google drive to convert the file to
	sf.file, err = srv.Files.
		Create(&file).
		Media(
			sourceData,                      // source file data
			googleapi.ContentType(xlsxMIME), // source file MIME type
		).
		Do()
	if err != nil {
		return "", err
	}
	
	return sf.file.Id, err
}

// generateName generates a (h) temporary filename
func generateName(prefix string) string {
	epoch := time.Now().Unix()
	return fmt.Sprintf("%s%d.xlsx", prefix, epoch)
}
