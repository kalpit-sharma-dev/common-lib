package app

import (
	"os"

	"gopkg.in/urfave/cli.v1"
)

//Command - Default command to find binary version
const Command = "Version"

//Execute is a function to be invoked, if default parameters are passed while plugin execution
type Execute func(args []string) error

//Create - Create an application to support both version and default execution of a binary
func Create(execute Execute) error {
	metadata, _ := GetMetadata() //nolint Ignoring Error as we want to initialize Application at any cost
	app := cli.NewApp()
	app.Name = metadata.StringFileInfo.Filename
	app.Version = metadata.StringFileInfo.ProductVersion
	app.Compiled = metadata.CompiledOn
	app.Copyright = metadata.StringFileInfo.Copyright
	app.Usage = metadata.StringFileInfo.Description
	app.HideHelp = false
	app.HideVersion = false

	app.Commands = []cli.Command{
		cli.Command{
			Name:            Command,
			Aliases:         []string{Command},
			Category:        metadata.StringFileInfo.ProductName,
			Usage:           "Binary Version Information",
			SkipFlagParsing: false,
			HideHelp:        false,
			Hidden:          false,
			Action: func(ctx *cli.Context) error {
				stream := os.Stdout
				defer stream.Close() //nolint
				return WriteVersion(stream)
			},
		},
	}

	app.Action = func(ctx *cli.Context) error {
		return execute(ctx.Args())
	}

	return app.Run(os.Args)
}
