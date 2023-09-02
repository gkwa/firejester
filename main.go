package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	git "github.com/go-git/go-git/v5"
	giturls "github.com/whilp/git-urls"
)

type Timestamp struct {
	EpochTime int64
	Time      string
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
	var extension string

	flag.StringVar(&repoPath, "path", "", "Path to the Git repository")
	flag.StringVar(&extension, "ext", "tar.xz", "Extension to name file")
	flag.Parse()

	if repoPath == "" {
		fmt.Println("Error: Please provide a valid path to the Git repository using the --path flag.")
		os.Exit(1)
	}

	if extension == "" {
		fmt.Println("Error: Please provide a extension string to name archive file")
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

	u, err := giturls.Parse(remoteURL)
	if err != nil {
		fmt.Printf("debug: %s", u.Scheme)
		panic(err)
	}

	remoteURL = fmt.Sprintf("%s:%s", u.Hostname(), u.Path)

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
		EpochTime: localTime.Unix(), // Epoch time in seconds
		Time:      localTime.Format(time.RFC3339),
	}

	fileName := fmt.Sprintf("%s_%d.%s", dirName, timestamp.EpochTime, extension)

	// Create a GitInfo struct.
	info := GitInfo{
		SHA:      sha.String(),
		ShortSHA: shortSHA,
		FileName: fileName,
		Repo:     remoteURL,
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

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	manifestPath := filepath.Join(cwd, fmt.Sprintf("manifest_%d.json", timestamp.EpochTime))

	// Create and write to a JSON file
	file, err := os.Create(manifestPath)
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
	f, err := os.OpenFile(manifestPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}
	if _, err := f.Write([]byte("\n")); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", manifestPath)
}
