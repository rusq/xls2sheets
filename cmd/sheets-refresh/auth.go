package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shibukawa/configdir"
	"golang.org/x/oauth2"
)

const (
	vendor      = "rusq"
	application = "sheets-refresh"

	tokFile = application + "-token.json"
)

// config directories
var configDirs = configdir.New(vendor, application)

type tokenMgr struct {
	Token *oauth2.Token
	file  string
}

func newMgrFromFile(filename string) (*tokenMgr, error) {
	mgr := new(tokenMgr)
	if err := mgr.Load(filename); err != nil {
		return nil, err
	}
	return mgr, nil
}

// Load creates a new token manager from token file.
func (mgr *tokenMgr) Load(filename string) error {
	// get the token from local storage
	tokenPath := mgr.path(filename)
	if tokenPath == "" {
		return fmt.Errorf("not found: %s", filename)
	}

	f, err := os.Open(tokenPath)
	if err != nil {
		return err
	}
	defer f.Close()

	token := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(&mgr.Token); err != nil {
		return err
	}

	mgr.Token = token
	mgr.file = tokenPath

	return nil
}

func (mgr *tokenMgr) Save(filename string) error {
	var fullPath = mgr.file
	if fullPath == "" {
		fullPath = mgr.createPath(filename)
	}

	log.Printf("Saving token file to: %s\n", fullPath)
	f, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	mgr.file = fullPath
	return json.NewEncoder(f).Encode(mgr.Token)
}

// createPath creates the path to token and returns the full path to
// tokenFile including tokenfilename.  I.e. on mac:
//    /Users/Youruser/Library/Caches/rusq/sheets-refresh/token.json
func (mgr *tokenMgr) createPath(path string) string {

	tokenPath := mgr.path(path)
	if tokenPath != "" {
		// do nothing if the path exists
		return tokenPath
	}

	cache := configDirs.QueryCacheFolder()
	if err := cache.MkdirAll(); err != nil {
		log.Fatalf("unable to create cache directory structure")
	}
	return filepath.Join(cache.Path, path)
}

func (mgr *tokenMgr) consoleRequest(config *oauth2.Config) error {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n\n"+
		"Enter authorization code: ", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	mgr.Token = tok

	return nil
}

// Retrieve a token, saves the token, then returns the generated client.
func googleClient(config *oauth2.Config) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	mgr, err := newMgrFromFile(tokFile)
	if err != nil {
		if err := mgr.consoleRequest(config); err != nil {
			return nil, fmt.Errorf("error requesting access token: %s", err)
		}
		if err := mgr.Save(tokFile); err != nil {
			return nil, fmt.Errorf("error saving token: %s", err)
		}
	}
	return config.Client(context.Background(), mgr.Token), nil
}

func (mgr *tokenMgr) Reset() {
	mgr.Token = nil
}

// removeToken finds and removes tokenFile from cache folder.  If the token
// file is not present it does nothing.
func removeToken(tokenFile string) error {
	tokenPath := getTokenPath(tokenFile)
	if tokenPath == "" {
		return nil
	}
	if err := os.Remove(tokenPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (tokenMgr) path(filename string) string {
	configDirs.LocalPath, _ = filepath.Abs(".")
	folder := configDirs.QueryFolderContainsFile(filename)
	if folder != nil {
		return filepath.Join(folder.Path, filename)
	}
	// check the existance in cache folder
	cache := configDirs.QueryCacheFolder()
	if cache.Exists(tokFile) {
		return filepath.Join(cache.Path, tokFile)
	}
	return ""
}

// getTokenPath returns the path for the tokenFile.  If file not found
// returns an empty string.
func getTokenPath(tokenFile string) string {
	// check the file locally and in user/system configuration folders
	configDirs.LocalPath, _ = filepath.Abs(".")
	folder := configDirs.QueryFolderContainsFile(tokenFile)
	if folder != nil {
		return filepath.Join(folder.Path, tokenFile)
	}
	// check the existance in cache folder
	cache := configDirs.QueryCacheFolder()
	if cache.Exists(tokFile) {
		return filepath.Join(cache.Path, tokFile)
	}
	return ""
}
