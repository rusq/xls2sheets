package xls2sheets

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"google.golang.org/api/sheets/v4"
)

// sheetSvc is a type that has a number of wrappers around subset of sheets
// Service functions.
type sheetSvc struct {
	svc           *sheets.Service
	spreadsheetID string
}

func newSheetSvc(client *http.Client, spreadsheetID string) (*sheetSvc, error) {
	svc, err := sheets.New(client)
	if err != nil {
		return nil, err
	}
	return &sheetSvc{svc: svc, spreadsheetID: spreadsheetID}, nil
}

// get returns a range of values from spreadsheet.
func (s *sheetSvc) get(Range string) (*sheets.ValueRange, error) {
	return s.svc.Spreadsheets.Values.Get(s.spreadsheetID, Range).Do()
}

// clear clears range within the target spreadsheet.
func (s *sheetSvc) clear(Range string) (*sheets.ClearValuesResponse, error) {
	// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/clear
	rb := &sheets.ClearValuesRequest{}
	return s.svc.Spreadsheets.Values.Clear(s.spreadsheetID, Range, rb).Do()
}

// addSheet adds a sheet.
func (s *sheetSvc) addSheet(address string) error {
	titleRange := strings.SplitN(address, "!", 2)
	if titleRange[0] == "" {
		return fmt.Errorf("invalid address: %q", address)
	}

	requests := []*sheets.Request{
		{AddSheet: &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{Title: titleRange[0]},
		}},
	}

	rb := &sheets.BatchUpdateSpreadsheetRequest{Requests: requests}

	_, err := s.svc.Spreadsheets.BatchUpdate(s.spreadsheetID, rb).Do()
	if err != nil {
		return err
	}
	return nil
}

func (s *sheetSvc) update(data *sheets.ValueRange) (*sheets.BatchUpdateValuesResponse, error) {
	const valueInputOption = userEntered // proper formatting of resulting values

	// Reference: https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/batchUpdate
	rb := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: valueInputOption,
		Data:             []*sheets.ValueRange{data},
	}

	resp, err := s.svc.Spreadsheets.Values.
		BatchUpdate(s.spreadsheetID, rb).
		Context(context.TODO()).
		Do()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *sheetSvc) validate(sheets []string, create bool) (*sheets.Spreadsheet, error) {
	// getting information about the spreadsheet
	log.Printf("  * retrieving information about the spreadsheet")
	spreadsheet, err := s.svc.Spreadsheets.Get(s.spreadsheetID).Do()
	if err != nil {
		return nil, err
	}

	log.Printf("  * validating target configuration")
	// need to ensure that all provided addresses are referencing valid
	// sheets
	for _, address := range sheets {
		valid := false

		for _, existing := range spreadsheet.Sheets {
			if strings.HasPrefix(address, existing.Properties.Title) {
				valid = true
				break
			}
		}
		if valid {
			continue
		}
		if !valid && create {
			if err := s.addSheet(address); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("address %q referencing nonexisting sheet - create it and restart", address)
		}
	}
	return spreadsheet, nil
}
