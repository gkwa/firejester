package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	git "github.com/go-git/go-git/v5"
)

type Timestamp struct {
	EpochTime    int64
	TimeWithZone string
}

// GitInfo represents the information about a Git repository.
type GitInfo struct {
	SHA      string
	ShortSHA string
	FileName string
}

func main() {
	// Define a command-line flag for the Git repository path.
	repoPath := flag.String("path", "", "Path to the Git repository")
	flag.Parse()

	// Ensure the --path flag is provided.
	if *repoPath == "" {
		fmt.Println("Error: Please provide a valid path to the Git repository using the --path flag.")
		os.Exit(1)
	}

	// Open the Git repository.
	r, err := git.PlainOpen(*repoPath)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Get the absolute path of the current directory
    absPath, err := filepath.Abs(*repoPath)
    if err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }

    // Extract the basename of the directory
    dirName := filepath.Base(absPath)

    fmt.Println("Absolute Path:", absPath)
    fmt.Println("Directory Name:", dirName)

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
		repoDir := filepath.Base(*repoPath)
		remoteURL = filepath.ToSlash(repoDir) // Convert to URL-friendly format
	}

	fmt.Println("Remote origin URL:", remoteURL)

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

	fileName := fmt.Sprintf("archive_%d.tgz", timestamp.EpochTime)

	// Create a GitInfo struct.
	info := GitInfo{
		SHA:      sha.String(),
		ShortSHA: shortSHA,
		FileName: fileName,
	}

	// Print the GitInfo struct.
	fmt.Println("Full SHA:", info.SHA)
	fmt.Println("Short SHA:", info.ShortSHA)
	fmt.Println("File Name:", info.FileName)

	// Marshal the struct to an indented JSON format
	jsonData, err := json.MarshalIndent(info, "", "    ") // Indent with four spaces
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
}
