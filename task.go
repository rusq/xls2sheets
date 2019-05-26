package xls2sheets

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/go-yaml/yaml"
)

// RefreshJob is a set of RefreshTasks
type RefreshJob struct {
	tasks refreshTasks

	sortedNames []string
}

type refreshTasks map[string]*RefreshTask

// FromConfig instantiates Job from config
func FromConfig(config []byte) (*RefreshJob, error) {
	tasks := make(refreshTasks)
	if err := yaml.Unmarshal(config, &tasks); err != nil {
		return nil, err
	}

	// Sorting job names
	names := make([]string, 0, len(tasks))
	for k := range tasks {
		names = append(names, k)
	}
	sort.Strings(names)

	job := &RefreshJob{
		tasks:       tasks,
		sortedNames: names,
	}

	return job, nil
}

// Execute executes the job. Tasks are run in alphabetical order.
// if any error occurs - the job is interrupted.
func (job *RefreshJob) Execute(client *http.Client) error {
	for _, taskName := range job.sortedNames {
		log.Printf("starting task: %q", taskName)
		if err := job.tasks[taskName].Run(client); err != nil {
			return fmt.Errorf("task %q: error: %s", taskName, err)
		}
		log.Printf("task %q: success", taskName)
	}

	return nil
}
