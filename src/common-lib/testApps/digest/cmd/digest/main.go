package main

import (
	"flag"
	"fmt"
	pkg "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/testApps/digest"
	"log"
	"os"
)

func main() {
	flag.Parse()

	if help {
		printUsage()
		os.Exit(0)
	}

	if create && verify {
		pErr("Choose -c (create) or -v (verify)")
	}

	if create {
		if file == "" {
			pErr("Input file is not set")
		}
		if digest == "" {
			digest = file + ext
		}
		if err := pkg.CreateSHA256Digest(file, digest); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if verify {
		if file == "" {
			pErr("The file for verification is not set")
		}
		if digest == "" {
			digest = file + ext
		}

		ok, err := pkg.CheckSHA256Digest(file, digest)
		if err != nil {
			log.Fatal(err)
		}
		if !ok {
			fmt.Println("Verification failed")
		} else {
			fmt.Println("Verification successful")
		}
	}
}

var (
	help         bool
	create       bool
	verify       bool
	file, digest string
	ext          string
)

func init() {
	flag.BoolVar(&help, "h", false, "print this help")
	flag.BoolVar(&create, "c", true, "create a digest")
	flag.BoolVar(&verify, "v", false, "verify the file against the digest")
	flag.StringVar(&file, "f", "", "input file")
	flag.StringVar(&digest, "d", "", "digest file")
	flag.StringVar(&ext, "e", ".sha256", "digest file extension")
}

func pErr(msg string) {
	fmt.Fprintf(os.Stderr, "ERROR: %s. Use -h to get help\n", msg)
	os.Exit(1)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}
