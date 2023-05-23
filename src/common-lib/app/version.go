package app

import (
	"encoding/json"
	"io"
	"time"
)

//Version - A struct to hold all the version related information about a binary
type Version struct {
	Name           string    `json:"name" cql:"component_name"`
	FileVersion    string    `json:"file_version" cql:"component_version"`
	ProductVersion string    `json:"product_version" cql:"product_version"`
	LastModifiedOn time.Time `json:"lastModifiedOn" cql:"lastmodifiedon"`
	BuildSHA       string    `json:"build_sha" cql:"build_sha"`
	CompiledOn     time.Time `json:"compiled_on" cql:"compiled_on"`
	RunningSince   time.Time `json:"running_since" cql:"running_since"`
}

//CreateVersion - A function to create a Version object from the binary Metadata
func CreateVersion() (Version, error) {
	metadata, err := GetMetadata()
	if err != nil {
		return Version{}, err
	}
	return Version{
		Name:           metadata.StringFileInfo.Filename,
		FileVersion:    metadata.FixedFileInfo.FileVersion.String(),
		ProductVersion: metadata.StringFileInfo.ProductVersion,
		BuildSHA:       metadata.BuildSHA,
		RunningSince:   metadata.RunningSince,
		CompiledOn:     metadata.CompiledOn,
	}, nil
}

//WriteVersion - A function to write version information on the Standard output stream
func WriteVersion(writer io.Writer) error {
	v, err := CreateVersion()
	if err != nil {
		return err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = writer.Write(b)
	if err != nil {
		return err
	}
	return nil
}
