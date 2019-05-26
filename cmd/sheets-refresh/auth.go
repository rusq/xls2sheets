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

	tokFile = "sheet-refresh-token.json"
)

// config directories
var configDirs = configdir.New(vendor, application)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
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
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	// get the token from local storage
	tokenPath := getTokenPath(file)
	if tokenPath == "" {
		return nil, fmt.Errorf("not found: %s", file)
	}

	f, err := os.Open(tokenPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(tokenFile string, token *oauth2.Token) {
	fullPath := createTokenPath(tokenFile)

	fmt.Printf("Saving credential file to: %s\n", fullPath)
	f, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// createTokenPath creates the path to token and returns the full path to
// tokenFile including tokenfilename.  I.e. on mac:
//    /Users/Youruser/Library/Caches/rusq/sheets-refresh/token.json
func createTokenPath(path string) string {
	//
	tokenPath := getTokenPath(path)
	if tokenPath != "" {
		return tokenPath
	}

	cache := configDirs.QueryCacheFolder()
	if err := cache.MkdirAll(); err != nil {
		log.Fatalf("unable to create cache directory structure")
	}
	return filepath.Join(cache.Path, path)
}

// removeToken finds and removes tokenFile from cache folder.  If the token
// file is not present it does nothing.
func removeToken(tokenFile string) error {
	tokenPath := getTokenPath(tokenFile)
	if tokenPath == "" {
		return nil
	}
	return os.Remove(tokenPath)
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
