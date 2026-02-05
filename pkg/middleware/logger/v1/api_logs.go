package loggerV1

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	Separator     *log.Logger
)

func CreateInitialFolder() {
	initLogFolder := "./system/"
	CreateDirectory(initLogFolder)
}

func CreateDirectory(path string) error {
	newPath := fmt.Sprintf("./logs/%s/", path)

	// Check if the directory already exists
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		// Create the directory with permissions 0755
		err := os.MkdirAll(newPath, 0755)
		if err != nil {
			return err
		}
		return nil
	} else {
		return err
	}
}

func SystemLogger(class, folder, filename, process, status string, request, response any) {
	// Checking folder name if exists
	currentTime := time.Now()
	folderName := folder + "/" + currentTime.Format("01-January")
	CreateDirectory(folderName)
	file, filErr := os.OpenFile(folderName+"/"+strings.ToLower(filename)+"-"+currentTime.Format("01022006")+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if filErr != nil {
		fmt.Println("error wrting the logs:", filErr.Error())
	}

	strRequest, _ := json.Marshal(request)
	strResponse, _ := json.Marshal(response)

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime)
	Separator := log.New(file, "", log.Ldate|log.Ltime)

	Separator.Println("")
	InfoLogger.Printf("%s | %s | %s\n", class, process, status)
	InfoLogger.Printf("Request: %s\n", string(strRequest))
	InfoLogger.Printf("Response: %s\n", string(strResponse))

	fmt.Printf("New entry for %s: %v\n", strings.ToUpper(process), currentTime.Format(time.DateTime))
	file.Close()
}

func SystemErrorLogger(class, folder, filename, process, status string, request, response any) {
	currentTime := time.Now()
	toLowerFilename := strings.ToLower(filename)
	folderLogPath := fmt.Sprintf("/error/%s/%s_%s", folder, toLowerFilename, currentTime.Format("01-January"))
	CreateDirectory(strings.ToLower(folderLogPath))
	file, filErr := os.OpenFile(strings.ToLower(folderLogPath)+"/"+strings.ToLower(filename)+"-"+currentTime.Format("01022006")+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if filErr != nil {
		fmt.Println("error wrting the logs:", filErr.Error())
	}

	strRequest, _ := json.Marshal(request)
	strResponse, _ := json.Marshal(response)

	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime)
	Separator := log.New(file, "", log.Ldate|log.Ltime)

	Separator.Println("")
	ErrorLogger.Printf("%s | %s | %s\n", class, process, status)
	ErrorLogger.Printf("Request: %s\n", string(strRequest))
	ErrorLogger.Printf("Response: %s\n", string(strResponse))

	fmt.Printf("New entry for %s: %v\n", strings.ToUpper(process), currentTime.Format(time.DateTime))
	file.Close()
}
