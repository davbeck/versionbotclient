package main

import "github.com/mkideal/cli"
import "os"
import "github.com/DHowett/go-plist"

type argT struct {
	cli.Helper
	Notation      string `cli:"n,notation" usage:"The version notation to display the version number as." dft:"dot"`
	InfoPlistPath string `cli:"p,info-plist" usage:"The path to the Info.plist file to get the display version of the app."`
	ShortVersion  string `cli:"v,short_version" usage:"The short user facing version number, such as 1.2.0. This overrides --info-plist."`
	Identifier    string `cli:"*i,identifier" usage:"The identifier for the app, for instance com.example.app."`
	TagGit        bool   `cli:"t,tag-git" usage:"Create a git tag with the new version." dft:"true"`
	PushGitTag    bool   `cli:"push-tag" usage:"When tagging the git repo with the new version number, this will push the new tag to origin when true." dft:"true"`
	HeaderPath    string `cli:"header" usage:"The path of a c header file to write the version information to."`
	BumpVersion   bool   `cli:"b,bump" usage:"When true the version number will be bumped by 1. This performs a network request and probably should only be used for archive builds."`
}

func (argv argT) versionName() string {
	if argv.ShortVersion != "" {
		return argv.ShortVersion
	}

	f, err := os.Open(argv.InfoPlistPath)
	if err != nil {
		panic(err)
	}

	decoder := plist.NewDecoder(f)
	var infoPlist interface{}
	decoder.Decode(&infoPlist)

	return infoPlist.(map[string]interface{})["CFBundleShortVersionString"].(string)
}

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		client := newVersionBotClient(argv.Identifier, argv.versionName(), argv.Notation)

		if argv.BumpVersion {
			client.bumpVersion()

			if argv.TagGit {
				if err := client.createGitTag(); err != nil {
					ctx.String("could not create git tag: %v\n", err)
				}

				if argv.PushGitTag {
					if err := client.pushGitTag(); err != nil {
						ctx.String("could not push git tag: %v\n", err)
					}
				}
			}
		} else {
			err := client.useGitVersion()
			if err != nil {
				ctx.String("could not get git version: %v\n", err)
			}

			ctx.String("configuration: %v\n", os.Getenv("CONFIGURATION"))
		}

		if argv.HeaderPath != "" {
			if err := client.writeHeader(argv.HeaderPath); err != nil {
				ctx.String("could not write version header: %v\n", err)
			}
		}

		ctx.String("version: %v\n", client.version)

		return nil
	})
}

// versionbotclient --info-plist testdata/Info.plist --identifier com.example.app --header tmp/header.h
