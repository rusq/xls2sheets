package xls2sheets

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"

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

// Update updates the target spreadsheet from source spreadsheet.
func (trg *Target) Update(client *http.Client, srcSheetID string, srcAddressRange []string) error {
	log.Printf("updating data in target spreadsheet %s", trg.SpreadsheetID)

	// TODO: copy everything from spreadsheet if sheetAddressRange and ts.SheetAddress is nil.
	if len(srcAddressRange) == 0 || len(trg.SheetAddress) == 0 {
		return errEmptyRange
	}
	if len(srcAddressRange) != len(trg.SheetAddress) {
		return errLengthMismatch
	}

	sheetsService, err := sheets.New(client)
	if err != nil {
		return err
	}
	sourcer := sheetSvc{svc: sheetsService, spreadsheetID: srcSheetID}
	updater := sheetSvc{svc: sheetsService, spreadsheetID: trg.SpreadsheetID}

	// validation of SheetAddresses
	if _, err := updater.validate(trg.SheetAddress, trg.Create); err != nil {
		return err
	}

	for sheetIdx := range srcAddressRange {
		log.Printf("  * copy range %q to %q", srcAddressRange[sheetIdx], trg.SheetAddress[sheetIdx])
		// getting source values
		values, err := sourcer.get(srcAddressRange[sheetIdx])
		if err != nil {
			return err
		}
		values.Range = trg.SheetAddress[sheetIdx]
		if trg.Clear {
			// clearing the spreadsheet
			log.Print("    * clearing target sheet")
			if _, err := updater.clear(trg.SheetAddress[sheetIdx]); err != nil {
				return err
			}
		}
		resp, err := updater.update(values)
		if err != nil {
			return err
		}
		log.Printf("    * OK: %d cells updated", resp.TotalUpdatedCells)
	}

	trg.Location = os.ExpandEnv(trg.Location)
	if trg.Location != "" {
		//save the file if location is set
		log.Printf("  * exporting to %s", trg.Location)
		if err := trg.download(client); err != nil {
			log.Print("    * export FAILED")
			return err
		}
		log.Print("    * export OK")
	}

	return nil
}

// download downloads the spreadsheet.
func (trg *Target) download(client *http.Client) error {
	if trg.Location == "" {
		return errors.New("target location is empty")
	}
	drv, err := drive.New(client)
	if err != nil {
		return err
	}
	resp, err := drv.Files.
		Export(
			trg.SpreadsheetID,
			mime.TypeByExtension(filepath.Ext(trg.Location))).
		Download()
	if err != nil {
		return err
	}
	if err := backup(trg.Location); err != nil {
		return fmt.Errorf("error creating a backup: %s", err)
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
