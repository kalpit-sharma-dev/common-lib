package app

import (
	"encoding/json"
	"fmt"
	"time"
)

//SymanticVersion - Struct to hold Binary Version information
type SymanticVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
	Build int `json:"build"`
}

//String - Function to convert symantic version into a string value
func (s SymanticVersion) String() string {
	return fmt.Sprintf("%v.%v.%v.%v", s.Major, s.Minor, s.Patch, s.Build)
}

//FixedFileInfo - Struct to hold File, Product version and file type information
type FixedFileInfo struct {
	FileVersion    SymanticVersion `json:"FileVersion"`
	ProductVersion SymanticVersion `json:"ProductVersion"`
	FileType       string          `json:"FileType"`
}

//StringFileInfo - Struct to hold, binary information as shown in the properties
type StringFileInfo struct {
	Description    string `json:"FileDescription"`
	Copyright      string `json:"LegalCopyright"`
	Filename       string `json:"OriginalFilename"`
	ProductVersion string `json:"ProductVersion"`
	ProductName    string `json:"ProductName"`
}

//Metadata - Stract to hold all the information related to binary
type Metadata struct {
	FixedFileInfo  FixedFileInfo  `json:"FixedFileInfo"`
	StringFileInfo StringFileInfo `json:"StringFileInfo"`
	RunningSince   time.Time      `json:"RunningSince"`
	CompiledOn     time.Time      `json:"CompiledOn"`
	BuildSHA       string         `json:"BuildSHA"`
}

var data *Metadata

//GetMetadata - A function to create Global metadata object based on given version info
func GetMetadata() (*Metadata, error) {
	if data == nil {
		date, err := time.Parse(time.RFC1123Z, CompiledOn)
		if err != nil {
			fmt.Println(err)
			date = time.Now()
		}
		data = &Metadata{RunningSince: time.Now(), BuildSHA: BuildCommitSHA, CompiledOn: date}
		err = json.Unmarshal([]byte(VersionInfo), data)
		return data, err
	}
	return data, nil
}
