package xls2sheets

import (
	"log"
	"net/http"
	"sort"

	"github.com/go-yaml/yaml"
)

// Job is a set of RefreshTasks
type Job struct {
	Tasks Tasks

	sortedNames []string
}

type Tasks map[string]*Task

// FromConfig instantiates Job from config
func FromConfig(config []byte) (*Job, error) {
	tasks := make(Tasks)
	if err := yaml.Unmarshal(config, &tasks); err != nil {
		return nil, err
	}

	// Sorting job names
	names := make([]string, 0, len(tasks))
	for k := range tasks {
		names = append(names, k)
	}
	sort.Strings(names)

	job := &Job{
		Tasks:       tasks,
		sortedNames: names,
	}

	return job, nil
}

// Execute executes the job. Tasks are run in alphabetical order.
// if any error occurs - the job is interrupted.
func (job *Job) Execute(client *http.Client) error {
	for _, taskName := range job.sortedNames {
		log.Printf("starting task: %q", taskName)
		if err := job.Tasks[taskName].Run(client); err != nil {
			log.Printf("task %q: error: %s", taskName, err)
			continue
		}
		log.Printf("task %q: success", taskName)
	}

	return nil
}
