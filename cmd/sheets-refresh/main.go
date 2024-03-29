package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/rusq/xls2sheets"
	"github.com/rusq/xls2sheets/internal/authmgr"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

var (
	build   = ""
	version = ""
)

var exepath = filepath.Dir(os.Args[0])

// command line parameters
var (
	resetAuth = flag.Bool("reset", false, "deletes the locally stored token before execution\n"+
		"this will trigger reauthentication")
	jobConfig   = flag.String("job", "", "configuration `file` with job definition")
	consoleAuth = flag.Bool("console", false, "use text authentication prompts instead of opening browser")
	ver         = flag.Bool("version", false, "print program version and quit")

	defaultCredentialsFile = filepath.Join(exepath, ".refresh-credentials.json")
	credentials            = flag.String("auth", defaultCredentialsFile, "file with authentication data")
)

func mustStr(s string, err error) string {
	if err != nil {
		panic(err)
	}
	return s
}

func main() {
	flag.Parse()

	if *ver {
		fmt.Printf("%s (%s)", version, build)
		os.Exit(0)
	}

	_, err := os.Stat(*credentials)
	if err != nil {
		fmt.Printf(credentialsHowTo, *credentials)
		os.Exit(1)
	}

	opts := []authmgr.Option{
		authmgr.OptTryWebAuth(!*consoleAuth, "/", ""),
		authmgr.OptAppName("rusq", "sheets-refresh"),
		authmgr.OptUseIndexPage(true),
	}
	if *resetAuth {
		opts = append(opts, authmgr.OptResetAuth())
	}

	// check parameters
	if *jobConfig == "" {
		if *resetAuth {
			os.Exit(0) // exiting without error if we were asked to just reset
		}
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
