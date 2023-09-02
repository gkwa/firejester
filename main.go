package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
)

type Timestamp struct {
	EpochTime    int64
	TimeWithZone string
}

type GitInfo struct {
	SHA      string
	ShortSHA string
	FileName string
	Repo     string
}

type PackgeInfo struct {
	Timestamp
	GitInfo
}

func main() {
	var repoPath string

	flag.StringVar(&repoPath, "path", "", "Path to the Git repository")
	flag.Parse()

	if repoPath == "" {
		fmt.Println("Error: Please provide a valid path to the Git repository using the --path flag.")
		os.Exit(1)
	}

	r, err := git.PlainOpen(repoPath)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	dirName := filepath.Base(absPath)

	// Get the remote origin
	remotes, err := r.Remotes()
	if err != nil {
		fmt.Println("Error getting remotes:", err)
		os.Exit(1)
	}

	var remoteURL string

	if len(remotes) > 0 {
		for _, remote := range remotes {
			if remote.Config().Name == "origin" {
				remoteURL = remote.Config().URLs[0]
				break
			}
		}
	} else {
		// If there are no remotes, use the local directory name
		repoDir := filepath.Base(repoPath)
		remoteURL = filepath.ToSlash(repoDir) // Convert to URL-friendly format
	}

	u, err := url.Parse(remoteURL)
	if err != nil {
		panic(err)
	}

	// redact credentials
	urlRedacted := strings.ReplaceAll(remoteURL, u.User.String(), "")
	urlRedacted = strings.ReplaceAll(urlRedacted, "@", "")

	fmt.Println("Remote origin URL:", urlRedacted)
	remoteURL = urlRedacted

	// Get the HEAD reference.
	ref, err := r.Head()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Get the SHA of the HEAD.
	sha := ref.Hash()
	shortSHA := ref.Hash().String()[:7] // Take the first 7 characters of the full SHA

	// Get the current time with the local time zone
	localTime := time.Now()

	// Create a Timestamp struct
	timestamp := Timestamp{
		EpochTime:    localTime.Unix(), // Epoch time in seconds
		TimeWithZone: localTime.Format("2006-01-02 15:04:05 -0700 MST"),
	}

	fileName := fmt.Sprintf("%s_%d", dirName, timestamp.EpochTime)

	// Create a GitInfo struct.
	info := GitInfo{
		SHA:      sha.String(),
		ShortSHA: shortSHA,
		FileName: fileName,
		Repo:     urlRedacted,
	}

	pi := PackgeInfo{
		GitInfo:   info,
		Timestamp: timestamp,
	}

	// Marshal the struct to an indented JSON format
	jsonData, err := json.MarshalIndent(pi, "", "    ") // Indent with four spaces
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	fileName2 := fmt.Sprintf("manifest_%d.json", timestamp.EpochTime)

	// Create and write to a JSON file
	file, err := os.Create(fileName2)
	if err != nil {
		fmt.Println("Error creating JSON file:", err)
		return
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing JSON data to file:", err)
		return
	}

	// Apppend neline to manifest
	f, err := os.OpenFile(fileName2, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}
	if _, err := f.Write([]byte("\n")); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
}
