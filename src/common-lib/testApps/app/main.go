package main

import (
	"fmt"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/app"
)

func main() {
	app.VersionInfo = "{\"FixedFileInfo\": { " +
		"\"FileVersion\": {    \"Major\": 1,    \"Minor\": 0,    \"Patch\": 0,    \"Build\": 0  }, " +
		" \"ProductVersion\": {    \"Major\": 1,    \"Minor\": 0,    \"Patch\": 0,    \"Build\": 0  }, " +
		" \"FileType\": \"01\"},\"StringFileInfo\": {  \"FileDescription\": \"\",  " +
		"\"LegalCopyright\": \"@IT Management Platform\",  \"OriginalFilename\": \"platform-version-plugin.exe\", " +
		" \"ProductVersion\": \"1.0.#\",  \"ProductName\": \"ITSPlatform\"},\"VarFileInfo\": " +
		"{  \"Translation\": {    \"LangID\": \"0409\"  }}\r\n}\r\n"
	err := app.Create(execute)
	fmt.Println(err)
}

var execute = func(args []string) error {
	fmt.Println("Args : ", args)
	return nil
}
