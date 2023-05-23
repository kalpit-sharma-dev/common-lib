package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/cheekybits/genny/parse"
)

/*

  source | genny gen [-in=""] [-out=""] [-pkg=""] "KeyType=string,int ValueType=string,int"

*/

const (
	_ = iota
	exitcodeInvalidArgs
	exitcodeInvalidTypeSet
	exitcodeStdinFailed
	exitcodeGenFailed
	exitcodeGetFailed
	exitcodeSourceFileInvalid
	exitcodeDestFileFailed

	getAction = "get"
)

func execute(in, out, pkgName *string, args []string) {
	var prefix = "https://github.com/metabition/gennylib/raw/master/"

	if len(args) < 2 {
		usage()
		os.Exit(exitcodeInvalidArgs)
	}

	if strings.ToLower(args[0]) != "gen" && strings.ToLower(args[0]) != getAction {
		usage()
		os.Exit(exitcodeInvalidArgs)
	}

	// parse the typesets
	var setsArg = args[1]
	if strings.ToLower(args[0]) == getAction {
		setsArg = args[2]
	}
	typeSets, err := parse.TypeSet(setsArg)
	checkAndExit(exitcodeInvalidTypeSet, err)

	var outWriter io.Writer
	if len(*out) > 0 {
		var outFile *os.File
		err = os.MkdirAll(path.Dir(*out), 0700)
		checkAndExit(exitcodeDestFileFailed, err)
		outFile, err = os.Create(*out)
		checkAndExit(exitcodeDestFileFailed, err)
		defer func() {
			checkAndLog(outFile.Close())
		}()
		outWriter = outFile
	} else {
		outWriter = os.Stdout
	}

	switch {
	case strings.ToLower(args[0]) == getAction:
		if len(args) != 3 {
			fmt.Println("not enough arguments to get")
			usage()
			os.Exit(exitcodeInvalidArgs)
		}
		var (
			r *http.Response
			b []byte
		)
		r, err = http.Get(prefix + args[1])
		checkAndExit(exitcodeGetFailed, err)
		defer func() {
			checkAndLog(r.Body.Close())
		}()
		b, err = ioutil.ReadAll(r.Body)
		checkAndExit(exitcodeGetFailed, err)
		br := bytes.NewReader(b)
		err = gen(*in, *pkgName, br, typeSets, outWriter)
	case len(*in) > 0:
		var file *os.File
		file, err = os.Open(*in)
		checkAndExit(exitcodeSourceFileInvalid, err)
		defer func() {
			checkAndLog(file.Close())
		}()
		err = gen(*in, *pkgName, file, typeSets, outWriter)
	default:
		var source []byte
		source, err = ioutil.ReadAll(os.Stdin)
		checkAndExit(exitcodeStdinFailed, err)
		reader := bytes.NewReader(source)
		err = gen("stdin", *pkgName, reader, typeSets, outWriter)
	}

	// do the work
	checkAndExit(exitcodeGenFailed, err)
}

func usage() {
	_, err := fmt.Fprintln(os.Stderr, `usage: genny [{flags}] gen "{types}"

gen - generates type specific code from generic code.
get <package/file> - fetch a generic template from the online library and gen it.

{flags}  - (optional) Command line flags (see below)
{types}  - (required) Specific types for each generic type in the source
{types} format:  {generic}={specific}[,another][ {generic2}={specific2}]

Examples:
  Generic=Specific
  Generic1=Specific1 Generic2=Specific2
  Generic1=Specific1,Specific2 Generic2=Specific3,Specific4

Flags:`)
	if err != nil {
		checkAndExit(exitcodeStdinFailed, err)
	}
	flag.PrintDefaults()
}

// gen performs the generic generation.
func gen(filename, pkgName string, in io.ReadSeeker, typesets []map[string]string, out io.Writer) error {
	var output []byte
	var err error

	output, err = parse.Generics(filename, pkgName, in, typesets)
	if err != nil {
		return err
	}

	_, err = out.Write(output)
	return err

}
