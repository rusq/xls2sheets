package xls2sheets

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"google.golang.org/api/sheets/v4"
)

// TargetSpreadsheet bears the information about the target spreadsheet and
// address within it
type TargetSpreadsheet struct {
	// SpreadsheetID is the Google Spreadsheet ID
	// i.e. 1lqbZm_TCsqcOTvOHPjG2CvZ6PpmDtBg_6qe-J1I91sk
	SpreadsheetID string `yaml:"spreadsheet_id"`
	// TargetSheet specifies the start location within the target
	// Google Sheet for all corresponding SheetAddressRange that
	// are defined on the source.  Example:  [ Sheet2!B4, Sheet3!A1 ]
	SheetAddress []string `yaml:"address"`
	// Clear specifies if the process should delete all data from the
	// Target Sheet before updating
	Clear  bool `yaml:"clear,omitempty"`
	Create bool `yaml:"create,omitempty"`
}

var (
	errEmptyRange     = errors.New("empty source and/or target ranges")
	errLengthMismatch = errors.New("source and target ranges have different lengths")
)

func debugPrintout(valueRange *sheets.ValueRange) {
	for rowIdx := range valueRange.Values {
		for colIdx := range valueRange.Values[rowIdx] {
			fmt.Printf("%s,", valueRange.Values[rowIdx][colIdx])
		}
		fmt.Println()
	}
}

// clearSheet clears sheet within the target spreadsheet
func (ts *TargetSpreadsheet) clearSheet(sheetsService *sheets.Service, Range string) (*sheets.ClearValuesResponse, error) {
	// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/clear
	rb := &sheets.ClearValuesRequest{}
	return sheetsService.Spreadsheets.Values.Clear(ts.SpreadsheetID, Range, rb).Do()
}

func (ts *TargetSpreadsheet) addSheetOrFail(sheetsService *sheets.Service, address string) error {
	if !ts.Create {
		// creating sheets is forbidden
		return fmt.Errorf("address %q referencing nonexisting sheet - create it and restart", address)
	}
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

	_, err := sheetsService.Spreadsheets.BatchUpdate(ts.SpreadsheetID, rb).Do()
	if err != nil {
		return err
	}
	return nil
}

// Update updates the target spreadsheet from source spreadsheet
func (ts *TargetSpreadsheet) Update(client *http.Client, spreadsheetID string, sheetAddressRange []string) error {
	log.Printf("updating data in target spreadsheet %s", ts.SpreadsheetID)

	if len(sheetAddressRange) == 0 || len(ts.SheetAddress) == 0 {
		return errEmptyRange
	}
	if len(sheetAddressRange) != len(ts.SheetAddress) {
		return errLengthMismatch
	}
	sheetsService, err := sheets.New(client)
	if err != nil {
		return err
	}
	// validation of SheetAddresses
	if _, err := ts.validate(sheetsService); err != nil {
		return err
	}

	for sheetIdx := range sheetAddressRange {
		log.Printf("  * copy range %q to %q", sheetAddressRange[sheetIdx], ts.SheetAddress[sheetIdx])
		// getting source values
		values, err := sheetsService.Spreadsheets.Values.Get(spreadsheetID, sheetAddressRange[sheetIdx]).Do()
		if err != nil {
			return err
		}
		values.Range = ts.SheetAddress[sheetIdx]
		if ts.Clear {
			// clearing the spreadsheet
			log.Print("    * clearing target sheet")
			if _, err := ts.clearSheet(sheetsService, ts.SheetAddress[sheetIdx]); err != nil {
				return err
			}
		}
		resp, err := ts.updateSheet(sheetsService, values)
		if err != nil {
			return err
		}
		log.Printf("    * OK: %d cells updated", resp.TotalUpdatedCells)
	}

	return nil
}

// updateSheet updates only one sheet
func (ts *TargetSpreadsheet) updateSheet(sheetsService *sheets.Service, data *sheets.ValueRange) (*sheets.BatchUpdateValuesResponse, error) {
	const valueInputOption = "USER_ENTERED" // proper formatting of resulting values

	// Reference: https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/batchUpdate
	rb := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: valueInputOption,
		Data:             []*sheets.ValueRange{data},
	}

	resp, err := sheetsService.Spreadsheets.Values.
		BatchUpdate(ts.SpreadsheetID, rb).
		Context(context.TODO()).
		Do()

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// validate checks if all defined in the configuration sheets exist and
// returns the *sheet.Spreadsheet structure.
func (ts *TargetSpreadsheet) validate(sheetsService *sheets.Service) (*sheets.Spreadsheet, error) {
	// getting information about the spreadsheet
	log.Printf("  * retrieving information about the spreadsheet")
	spreadsheet, err := sheetsService.Spreadsheets.Get(ts.SpreadsheetID).Do()
	if err != nil {
		return nil, err
	}

	log.Printf("  * validating target configuration")
	// need to ensure that all provided addresses are referencing valid
	// sheets
	for _, address := range ts.SheetAddress {
		valid := false

		for _, sheet := range spreadsheet.Sheets {
			if strings.HasPrefix(address, sheet.Properties.Title) {
				valid = true
				break
			}
		}
		if !valid {
			if err := ts.addSheetOrFail(sheetsService, address); err != nil {
				return nil, err
			}

		}
	}
	return spreadsheet, nil
}
