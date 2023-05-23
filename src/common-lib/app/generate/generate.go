package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

//go:generate sh -c "go run ./generate.go > ../flag.go && go fmt ../flag.go"
func main() {
	dat, err := ioutil.ReadFile("versioninfo.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("package app")
	fmt.Println("")
	versionInfo := string(dat)
	versionInfo = strings.Replace(versionInfo, "\n", "", -1)
	versionInfo = strings.Replace(versionInfo, "\r", "", -1)
	versionInfo = strings.Replace(versionInfo, "\r\n", "", -1)
	versionInfo = strings.Replace(versionInfo, " ", "", -1)
	versionInfo = strings.Replace(versionInfo, "\"", "\\\"", -1)
	re := regexp.MustCompile(`\r?\n`)
	versionInfo = re.ReplaceAllString(versionInfo, "")
	fmt.Println("// VersionInfo - a variable to hold Version info Json created by Jenking Job")
	fmt.Println("// having patch and Build Numbers")
	fmt.Printf("var VersionInfo=\"%s\"", versionInfo)
	fmt.Println("")
	fmt.Println("")
	fmt.Println("//CompiledOn - A variable to hold Binary compilation Date")
	fmt.Printf("var CompiledOn=\"%s\"", time.Now().UTC().Format(time.RFC1123Z))
	fmt.Println("")
	fmt.Println("")
	fmt.Println("//BuildCommitSHA is a SHA id from the code base, while building this binary")
	fmt.Println("var BuildCommitSHA=\"\"")
}
