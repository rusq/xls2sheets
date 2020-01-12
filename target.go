package xls2sheets

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

const userEntered = "USER_ENTERED"

const (
	bakSuffix   = ".bak" // backup file suffix, will be added to file
	bakFileMode = 0666
)

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
func (trg *Target) clearSheet(sheetsService *sheets.Service, Range string) (*sheets.ClearValuesResponse, error) {
	// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/clear
	rb := &sheets.ClearValuesRequest{}
	return sheetsService.Spreadsheets.Values.Clear(trg.SpreadsheetID, Range, rb).Do()
}

func (trg *Target) addSheetOrFail(sheetsService *sheets.Service, address string) error {
	if !trg.Create {
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

	_, err := sheetsService.Spreadsheets.BatchUpdate(trg.SpreadsheetID, rb).Do()
	if err != nil {
		return err
	}
	return nil
}

// Update updates the target spreadsheet from source spreadsheet.
func (trg *Target) Update(client *http.Client, srcSheetID string, sheetAddressRange []string) error {
	log.Printf("updating data in target spreadsheet %s", trg.SpreadsheetID)

	// TODO: copy everything from spreadsheet if sheetAddressRange and ts.SheetAddress is nil.
	if len(sheetAddressRange) == 0 || len(trg.SheetAddress) == 0 {
		return errEmptyRange
	}
	if len(sheetAddressRange) != len(trg.SheetAddress) {
		return errLengthMismatch
	}

	trg.Location = os.ExpandEnv(trg.Location)

	sheetsService, err := sheets.New(client)
	if err != nil {
		return err
	}
	// validation of SheetAddresses
	if _, err := trg.validate(sheetsService); err != nil {
		return err
	}

	for sheetIdx := range sheetAddressRange {
		log.Printf("  * copy range %q to %q", sheetAddressRange[sheetIdx], trg.SheetAddress[sheetIdx])
		// getting source values
		values, err := sheetsService.Spreadsheets.Values.Get(srcSheetID, sheetAddressRange[sheetIdx]).Do()
		if err != nil {
			return err
		}
		values.Range = trg.SheetAddress[sheetIdx]
		if trg.Clear {
			// clearing the spreadsheet
			log.Print("    * clearing target sheet")
			if _, err := trg.clearSheet(sheetsService, trg.SheetAddress[sheetIdx]); err != nil {
				return err
			}
		}
		resp, err := trg.updateSheet(sheetsService, values)
		if err != nil {
			return err
		}
		log.Printf("    * OK: %d cells updated", resp.TotalUpdatedCells)
	}
	if trg.Location != "" {
		//save the file if location is set
		log.Printf("  * trying to export to %s", trg.Location)
		if err := trg.download(client); err != nil {
			log.Print("    * export FAILED")
			return err
		}
		log.Print("    * export OK")
	}

	return nil
}

// updateSheet updates only one sheet
func (trg *Target) updateSheet(sheetsService *sheets.Service, data *sheets.ValueRange) (*sheets.BatchUpdateValuesResponse, error) {
	const valueInputOption = userEntered // proper formatting of resulting values

	// Reference: https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/batchUpdate
	rb := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: valueInputOption,
		Data:             []*sheets.ValueRange{data},
	}

	resp, err := sheetsService.Spreadsheets.Values.
		BatchUpdate(trg.SpreadsheetID, rb).
		Context(context.TODO()).
		Do()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (trg *Target) download(client *http.Client) error {
	if trg.Location == "" {
		return errors.New("target location is empty")
	}
	if err := prepareFile(trg.Location); err != nil {
		return err
	}
	drv, err := drive.New(client)
	if err != nil {
		return err
	}
	resp, err := drv.Files.Export(trg.SpreadsheetID, mime.TypeByExtension(filepath.Ext(trg.Location))).Download()
	if err != nil {
		return err
	}
	f, err := os.Create(trg.Location)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}
	return nil
}

// prepareFile checks the filepath and removes the file if it exists
// (maybe would be good to make a backup).
func prepareFile(filename string) error {
	fi, err := os.Stat(filename)
	if err != nil && fi == nil {
		// no file
		return nil
	}
	if fi != nil && fi.IsDir() {
		return fmt.Errorf("%s is a directory, will not overwrite", filename)
	}
	if err := backup(filename); err != nil {
		return fmt.Errorf("error creating a backup: %s", err)
	}
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("unable to remove the previous version of local copy: %s", err)
	}
	return nil
}

func backup(filename string) error {
	bakFilename := filename + bakSuffix
	bak, err := os.Create(bakFilename)
	if err != nil {
		return fmt.Errorf("unable to overwrite backup file: %s", err)
	}
	defer bak.Close()
	src, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("unable to open the local file for backup: %s", err)
	}
	defer src.Close()
	if _, err := io.Copy(bak, src); err != nil {
		return fmt.Errorf("failed to make a backup: %s", err)
	}
	return nil
}

// validate checks if all defined in the configuration sheets exist and
// returns the *sheet.Spreadsheet structure.
func (trg *Target) validate(sheetsService *sheets.Service) (*sheets.Spreadsheet, error) {
	// getting information about the spreadsheet
	log.Printf("  * retrieving information about the spreadsheet")
	spreadsheet, err := sheetsService.Spreadsheets.Get(trg.SpreadsheetID).Do()
	if err != nil {
		return nil, err
	}

	log.Printf("  * validating target configuration")
	// need to ensure that all provided addresses are referencing valid
	// sheets
	for _, address := range trg.SheetAddress {
		valid := false

		for _, sheet := range spreadsheet.Sheets {
			if strings.HasPrefix(address, sheet.Properties.Title) {
				valid = true
				break
			}
		}
		if !valid {
			if err := trg.addSheetOrFail(sheetsService, address); err != nil {
				return nil, err
			}

		}
	}
	return spreadsheet, nil
}
