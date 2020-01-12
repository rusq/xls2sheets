package xls2sheets

import (
	"net/http"
)

// NewTask creates the task
func NewTask(source *Source, target *Target) *Task {
	t := &Task{
		Source: source,
		Target: target,
	}
	return t
}

// Run runs the refresh task
func (task *Task) Run(client *http.Client) error {
	// fetch from source and upload to google drive
	tempSpreadsheetID, err := task.Source.Process(client)
	if err != nil {
		return err
	}
	// this ensures that the temporary file is deleted at the end of
	// conversion
	if !task.LeaveJunk {
		defer task.Source.Delete(client)
	}
	// copy data from temporary file to target file
	if err := task.Target.Update(client, tempSpreadsheetID, task.Source.SheetAddressRange); err != nil {
		return err
	}
	return nil
}
