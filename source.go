package xls2sheets

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

const (
	// tempFilePrefix used for temporary file uploads
	tempFilePrefix = "xls2sheets$"

	minFileSz = 512 // can't imagine excel file smaller than this.
)

type srcType string

const (
	srcUnknown srcType = "<unknown>"
	srcDisk    srcType = "local file"
	srcWeb     srcType = "remote file"
	srcGSheet  srcType = "Google Spreadsheet"
)

var converters = map[srcType]sourcer{
	srcDisk:   disk{},
	srcWeb:    web{},
	srcGSheet: gsheet{},
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

	fileID string // uploaded file ID
}

// sourcer is the source file interface
type sourcer interface {
	// convert converts the source document to google sheets format and
	// returns the drive.fileID (same as sheetID)
	convert(client *http.Client, loc string) (sheetID string, err error)
}

// different source types
type disk struct{}
type web struct{}
type gsheet struct{}

var gsheetRe = regexp.MustCompile(`[-\w]{25,}$`)

// Errors.
var (
	errNothingToDelete = errors.New("delete called before upload")
	errTooSmall        = errors.New("file is too small, is that even a spreasheet")
	errUnknown         = errors.New("unknown file type or location")
)

func fileType(loc string) srcType {
	switch {
	default:
		return srcUnknown
	case strings.HasPrefix(strings.ToLower(loc), "file://"):
		return srcDisk
	case strings.Contains(loc, "://"):
		return srcWeb
	case gsheetRe.MatchString(loc):
		return srcGSheet
	case fileExists(loc):
		return srcDisk
	}
}

func fileExists(loc string) bool {
	_, err := os.Stat(loc)
	if err != nil {
		return false
	}
	return true
}

func getFilename(loc string) (string, error) {
	url, err := url.Parse(loc)
	if err != nil {
		return "", err
	}
	// dirty hax
	return filepath.Join(url.Host, url.Path), nil
}

// Process gets the file onto google drive, if needed (i.e. it not google
// spreadsheet).  Returns the file ID on google drive.
func (sf *SourceFile) Process(client *http.Client) (string, error) {
	sf.FileLocation = os.ExpandEnv(sf.FileLocation)
	// determine file type
	typ := fileType(sf.FileLocation)
	if typ == srcUnknown {
		return "", errUnknown
	}
	log.Printf("+ type detected as: %s", typ)

	// getting appropriate converter for the source type
	c, ok := converters[typ]
	if !ok {
		return "", errUnknown
	}

	log.Printf("+ trying to open: %s", sf.FileLocation)
	id, err := c.convert(client, sf.FileLocation)
	if err != nil {
		return "", err
	}

	// saving fileID, delete will need it.
	sf.fileID = id

	return id, nil
}

// Delete deletes the temporary file from the google drive.
func (sf *SourceFile) Delete(client *http.Client) error {
	// if the fileID is nil, then upload function hasn't been called yet
	if sf.fileID == "" {
		return errNothingToDelete
	}
	srv, err := drive.New(client)
	if err != nil {
		return err
	}
	if err := srv.Files.Delete(sf.fileID).Do(); err != nil {
		return err
	}
	// clearing the file ID so that consequent calls would now that the file
	// does not exist
	sf.fileID = ""
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
	hFile, err := srv.Files.
		Create(&file).
		Media(
			sourceData, // source file data
			googleapi.ContentType(mime.TypeByExtension(filepath.Ext(sf.FileLocation))), // source file MIME type
		).
		Do()
	if err != nil {
		return "", err
	}
	sf.fileID = hFile.Id

	return file.Id, err
}

// generateName generates a temporary filename to save on Google Drive.
func generateName(prefix string, extension string) string {
	epoch := time.Now().Unix()
	return fmt.Sprintf("%s%d%s", prefix, epoch, extension)
}

func (web) convert(client *http.Client, loc string) (string, error) {
	f, err := fetchFromWeb(loc)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return upload(client, f, loc)
}

func (disk) convert(client *http.Client, loc string) (string, error) {
	if strings.HasPrefix(strings.ToLower(loc), "file://") {
		var err error
		if loc, err = getFilename(loc); err != nil {
			return "", err
		}
	}
	fi, err := os.Stat(loc)
	if err != nil {
		return "", err
	}
	if fi.Size() < minFileSz {
		return "", errTooSmall
	}

	f, err := os.Open(loc)
	if err != nil {
		return "", err
	}
	defer f.Close()

	return upload(client, f, loc)
}

func (gsheet) convert(client *http.Client, loc string) (string, error) {
	return loc, nil
}

// upload uploads the source data to temporary google spreadsheet on
// google drive, so that it would be possible to copy data from it.
func upload(client *http.Client, sourceData io.Reader, srcName string) (string, error) {
	srv, err := drive.New(client)
	if err != nil {
		return "", err
	}
	// target file name and MIME type format, so that Google Drive would
	// convert the source excel file to Google Sheets format
	file := drive.File{
		Name:     generateName(tempFilePrefix, filepath.Ext(srcName)),
		MimeType: gsheetMIME,
	}
	// content type is necessary for google drive to convert the file to
	hFile, err := srv.Files.
		Create(&file).
		Media(
			sourceData, // source file data
			googleapi.ContentType(mime.TypeByExtension(filepath.Ext(srcName))), // source file MIME type
		).
		Do()
	if err != nil {
		return "", err
	}
	return hFile.Id, err
}

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
