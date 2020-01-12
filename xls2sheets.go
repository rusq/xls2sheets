package xls2sheets

import (
	"log"
	"net/http"
	"sort"

	"github.com/go-yaml/yaml"
)

// Job is a collection of Tasks
type Job struct {
	Tasks Tasks

	sortedNames []string // cache of sorted task names
}

// Tasks is a mapping of {taskName: Task}.
type Tasks map[string]*Task

// Task contains all information needed to refresh the Google
// Spreadsheet from an external file.
type Task struct {
	Source *Source `yaml:"source"` // Source file info (defined below)
	Target *Target `yaml:"target"` // Target sheet info (defined below)

	LeaveJunk bool `yaml:"leave_junk,omitempty"` // leave temporary files on google disk
}

// Source contains the information about the source file and
// address + range of cells to copy
type Source struct {
	// Location specifies the file location
	// Valid values:
	//
	// 		https://www.example.com/dataset.xlsx
	//		file://MyWorkbook.xlsx
	//      somefile.ods
	FileLocation string `yaml:"location"`
	// SheetAddress is the address within the source workbook.
	// I.e. "Data!A1:U"
	SheetAddressRange []string `yaml:"address_range"`

	fileID   string // temporary spreadsheet ID
	tempName string //temporary spreadsheet file name
}

// Target bears the information about the target spreadsheet and
// address within it
type Target struct {
	// SpreadsheetID is the Google Spreadsheet ID.
	// Example: 1lqbZm_TCsqcOTvOHPjG2CvZ6PpmDtBg_6qe-J1I91sk
	SpreadsheetID string `yaml:"spreadsheet_id"`
	// Location (optional) is the location of the exported file on local disk.
	// This will save the Google Spreadsheet to local disk.
	// Example: "/Users/Anna/Documents/rates.xlsx"
	Location string `yaml:"location,omitempty"`
	// TargetSheet specifies the start location within the target
	// Google Sheet for all corresponding SheetAddressRange that
	// are defined on the source.  Example:  [ Sheet2!B4, Sheet3!A1 ]
	SheetAddress []string `yaml:"address"`
	// Clear (optional) specifies if the process should delete all data from
	// the Target Sheet before updating.
	Clear bool `yaml:"clear,omitempty"`
	// Create (optional) specifies if the process should create worksheet
	// if it does not exist.
	Create bool `yaml:"create,omitempty"`
}

// NewJobFromConfig instantiates Job from config
func NewJobFromConfig(config []byte) (*Job, error) {
	tasks := make(Tasks)
	if err := yaml.Unmarshal(config, &tasks); err != nil {
		return nil, err
	}

	job := &Job{
		Tasks: tasks,
	}

	return job, nil
}

// TaskNames returns the alphabetically sorted slice of task names.
func (j *Job) TaskNames() []string {
	if len(j.sortedNames) == len(j.Tasks) {
		return j.sortedNames
	}

	// Sorting job names
	names := make([]string, 0, len(j.Tasks))
	for k := range j.Tasks {
		names = append(names, k)
	}
	sort.Strings(names)

	j.sortedNames = names

	return j.sortedNames
}

// Execute executes the job.  Tasks are ran in alphabetical order.
// if any error occurs - the job is interrupted.
func (j *Job) Execute(client *http.Client) error {
	if len(j.Tasks) == 0 {
		log.Println("job has no tasks, nothing to do")
		return nil
	}
	for _, taskName := range j.TaskNames() {
		log.Printf("starting task: %q", taskName)
		if err := j.Tasks[taskName].Run(client); err != nil {
			log.Printf("task %q: error: %s", taskName, err)
		} else {
			log.Printf("task %q: success", taskName)
		}
	}

	return nil
}
