package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/rusq/xls2sheets"
	"github.com/rusq/xls2sheets/authmgr"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

var build = ""

// command line parameters
var (
	resetAuth = flag.Bool("reset", false, "deletes the locally stored token before execution\n"+
		"this will trigger reauthentication")
	jobConfig   = flag.String("job", "", "configuration `file` with job definition")
	consoleAuth = flag.Bool("console", false, "use text authentication prompts instead of opening browser")
	version     = flag.Bool("version", false, "print program version and quit")

	credentials *string
)

func init() {
	home, _ := os.UserHomeDir()
	defaultCredentialsFile := filepath.Join(home, ".refresh-credentials.json")
	credentials = flag.String("auth", defaultCredentialsFile, "file with authentication data")
}

func main() {
	flag.Parse()

	if *version {
		fmt.Println(build)
		os.Exit(0)
	}

	_, err := os.Stat(*credentials)
	if err != nil {
		fmt.Printf(credentialsHowTo, *credentials)
		os.Exit(1)
	}

	// check parameters
	if *jobConfig == "" {
		log.Fatal("no -job <yaml file> specified")
	}

	// read the configuration file
	jobData, err := ioutil.ReadFile(*jobConfig)
	if err != nil {
		log.Fatal(err)
	}

	// initialise job from the configuration file data
	job, err := xls2sheets.NewJobFromConfig(jobData)
	if err != nil {
		log.Fatal(err)
	}

	opts := []authmgr.Option{
		authmgr.OptTryWebAuth(!*consoleAuth, ""),
		authmgr.OptAppName("rusq", "sheets-refresh"),
		authmgr.OptUseIndexPage(true),
	}
	if *resetAuth {
		opts = append(opts, authmgr.OptResetAuth())
	}

	// prepare config from provided credentials file
	mgr, err := authmgr.NewFromGoogleCreds(*credentials, []string{sheets.SpreadsheetsScope, drive.DriveScope}, opts...)
	if err != nil {
		log.Fatal(err)
	}

	// initialising client
	client, err := mgr.Client()
	if err != nil {
		log.Fatal(err)
	}

	// running job
	if err := job.Execute(client); err != nil {
		log.Fatal(err)
	}
}
