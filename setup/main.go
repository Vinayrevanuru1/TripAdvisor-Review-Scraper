package main

// Dependencies
import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
)

var (
	errDirectoryCreation = errors.New("FAILED TO CREATE DIRECTORY")
	errGetDirectory      = errors.New("FAILED TO GET THE CURRENT DIRECTORY")
	errCloneRepo         = errors.New("FAILED TO CLONE THE REPOSITORY")
)

func main() {

	// Get the current directory
	currentDir, err := getCurrentDir()

	// Check for errors
	errorHandler(err)

	// Print the current directory
	fmt.Println("1. Current directory: ", currentDir)

	// Create a temporary directory to hold the repository
	tmpDirName, err := createDirectory("tmp")

	// Check for errors
	errorHandler(err)

	tmpDirFullPath := filepath.Join(currentDir, tmpDirName)

	// Print the message
	fmt.Println("2. Tmp Directory created:", tmpDirFullPath)

	// Create a temporary directory to hold the repository
	projectDirName, err := createDirectory("Project_Files")

	// Check for errors
	errorHandler(err)

	projectDirFullPath := filepath.Join(currentDir, projectDirName)

	// Print the message
	fmt.Println("3. Project Directory created:", projectDirFullPath)

	// Call the clone repo function
	msg, err := cloneRepo(tmpDirFullPath)

	// Check for errors
	errorHandler(err)

	fmt.Println("4. " + msg)

	// Copy docker-compose-prod.yml to the Project_Files directory
	msg, err = copy(
		filepath.Join(tmpDirFullPath, "docker-compose-prod.yml"),
		filepath.Join(projectDirFullPath, "docker-compose-prod.yml"))

	// Check for errors
	errorHandler(err)
	fmt.Println("5. " + msg)
}

// The function to get the current working directory
func getCurrentDir() (string, error) {
	// Get the current directory
	pwd, err := os.Getwd()
	// Check for errors
	if err != nil {
		return "Failed", errGetDirectory
	}

	return pwd, err
}

// The function to create a directory
func createDirectory(name string) (string, error) {

	// Create the directory
	err := os.Mkdir(name, os.ModePerm)

	// Check for errors
	if err != nil {
		return "Failed to Create Directory", errDirectoryCreation
	}
	return name, nil
}

func cloneRepo(path string) (string, error) {
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:        "https://github.com/algo7/TripAdvisor-Review-Scraper.git",
		RemoteName: "origin",
		Progress:   os.Stdout,
	})

	if err != nil {
		return "Failed", errCloneRepo
	}

	return "Repo cloned", nil
}

func errorHandler(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}

func copy(sourceFile string, destFile string) (string, error) {

	// Read the source file
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return "Ops", err
	}

	// Write tp the destination file
	err = ioutil.WriteFile(destFile, input, os.ModePerm)
	if err != nil {
		return "Ops", err
	}

	return "docker-compose-prod.yml copied successfully", nil
}
