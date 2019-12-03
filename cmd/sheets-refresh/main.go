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

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

var build = ""

var defaultCredentialsFile = os.ExpandEnv(filepath.Join("${HOME}", ".refresh-credentials.json"))

// command line parameters
var (
	resetAuth = flag.Bool("reset", false, "deletes the locally stored token before execution\n"+
		"this will trigger reauthentication")
	credentials = flag.String("auth", defaultCredentialsFile, "file with authentication data")
	jobConfig   = flag.String("job", "", "configuration `file` with job definition")
	consoleAuth = flag.Bool("console", false, "use text authentication prompts instead of opening browser")
	version     = flag.Bool("version", false, "print program version and quit")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Println(build)
		os.Exit(0)
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
	job, err := xls2sheets.FromConfig(jobData)
	if err != nil {
		log.Fatal(err)
	}

	opts := []authmgr.Option{
		authmgr.OptTryWebAuth(!*consoleAuth, ""),
		authmgr.OptAppName("rusq", "sheets-refresh"),
		authmgr.OptUseIndexPage(true),
	}

	// prepare config from provided credentials file
	mgr, err := authmgr.NewFromGoogleCreds(*credentials, []string{sheets.SpreadsheetsScope, drive.DriveFileScope}, opts...)
	if err != nil {
		log.Fatal(err)
	}

	if *resetAuth {
		if err := mgr.RemoveToken(); err != nil {
			log.Fatal(err)
		}
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

// prepareConfig loads configuration from disk and prepares oauth2.Config
func prepareConfig(credentialsFile string) (*oauth2.Config, error) {
	fileInfo, err := os.Stat(credentialsFile)
	if err != nil {
		return nil, err
	}

	// check if the permissions on the file are set correctly
	permissions := fileInfo.Mode().Perm()
	if !(permissions == 0600 || permissions == 0400) {
		return nil, fmt.Errorf("credentials file is to permissive (%o), "+
			"to fix - run:\n\tchmod 600 %s", permissions, credentialsFile)
	}

	b, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	// scope reference: https://developers.google.com/identity/protocols/googlescopes
	config, err := google.ConfigFromJSON(
		b,
		sheets.SpreadsheetsScope, // we need to be able to edit existing files.
		drive.DriveScope,         // this type of access is required to export sheets to files
	)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file to config: %v", err)
	}
	return config, nil
}
