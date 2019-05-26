package xls2sheets

import (
	"net/http"
)

// RefreshTask contains all information needed to refresh the Google
// Spreadsheet from an external file
type RefreshTask struct {
	Source *SourceFile        `yaml:"source"` // Source file info (defined below)
	Target *TargetSpreadsheet `yaml:"target"` // Target sheet info (defined below)
}

// NewTask creates the task
func NewTask(source *SourceFile, target *TargetSpreadsheet) *RefreshTask {
	return &RefreshTask{
		Source: source,
		Target: target,
	}
}

// Run runs the refresh task
func (task *RefreshTask) Run(client *http.Client) error {
	// fetch from source and upload to google drive
	tempSpreadsheetID, err := task.Source.FetchAndUpload(client)
	if err != nil {
		return err
	}
	// this ensures that the temporary file is deleted at the end of
	// conversion
	defer task.Source.Delete(client)
	// copy data from temporary file to target file
	if err := task.Target.Update(client, tempSpreadsheetID, task.Source.SheetAddressRange); err != nil {
		return err
	}
	return nil
}