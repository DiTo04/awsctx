package main

import (
	"bufio"
	"fmt"
	"github.com/dtennander/awsctx/awsctx"
	"github.com/urfave/cli"
	"log"
	"os"
	"sort"
	"strings"
)

const currentContextColor = "\033[1;33m"
const normalColor = "\033[0m"

var awsFolder string
var nameFlag string
var noColorFlag = false

func main() {
	app := cli.NewApp()
	app.Name = "awsctx"
	app.Version = "1.2"
	app.HideVersion = true

	app.HelpName = "awsctx"
	app.Usage = "A tool to switch aws profiles"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "folder, f",
			Usage:       "aws folder",
			Destination: &awsFolder,
			Value:       os.Getenv("HOME") + "/.aws",
		},
		cli.BoolFlag{
			Name:        "no-color, nc",
			Usage:       "remove color from output",
			Destination: &noColorFlag,
		},
	}

	app.Action = mainAction
	app.Commands = cli.Commands{
		{
			Name:        "rename",
			ArgsUsage:   "<old name> <new name>",
			Description: "Rename a profile to a new name.",
			Usage:       "renames a profile to a new name",
			ShortName:   "r",
			Action:      rename,
		}, {
			Name:        "-",
			Description: "Switch to the previous profile",
			Usage:       "switch to the previous profile",
			Action:      switchBack,
		}, {
			Name:   "<profile>",
			Usage:  "switch to given profile",
			Action: nil,
		}, {
			Name:   "version",
			Usage:  "prints the current version",
			Action: printVersion,
		}, {
			Name: "current",
			Usage: "prints the current profile",
			Action: printProfile,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func rename(c *cli.Context) error {
	if c.NArg() != 2 {
		return cli.NewExitError("Expected old name and new name.", 1)
	}
	aws, err := initAwsctx()
	if err != nil {
		return err
	}
	return aws.RenameProfile(c.Args()[0], c.Args()[1])
}

func setup() error {
	scanner := bufio.NewReader(os.Stdin)
	print("Please name your current default profile: ")
	input, err := scanner.ReadString('\n')
	if err != nil {
		return err
	}
	name := strings.TrimRight(input, "\n")
	err = awsctx.SetUpDefaultContext(awsFolder, name)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return err
}

func switchBack(c *cli.Context) error {
	aws, err := initAwsctx()
	if err != nil {
		return err
	}
	return aws.SwitchBack()
}

func mainAction(c *cli.Context) error {
	aws, err := initAwsctx()
	if err != nil {
		return err
	}
	switch c.NArg() {
	case 0:
		return printAllProfiles(aws)
	case 1:
		return aws.SwitchProfile(c.Args()[0])
	default:
		return cli.NewExitError("expected one or zero arguments", 1)
	}
}

func initAwsctx() (*awsctx.Awsctx, error) {
	ctx, err := awsctx.New(awsFolder)
	if err != nil {
		_, ok := err.(awsctx.NoContextError)
		if !ok {
			return nil, err
		}
		err := setup()
		if err != nil {
			return nil, err
		}
		return initAwsctx()
	}
	return ctx, err
}

func printAllProfiles(aws *awsctx.Awsctx) error {
	profiles, err := aws.GetProfiles()
	if err != nil {
		return err
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })
	for _, profile := range profiles {
		var prefix string
		switch {
		case noColorFlag && profile.IsCurrent:
			prefix = "*"
		case !noColorFlag && profile.IsCurrent:
			prefix = currentContextColor
		case noColorFlag && !profile.IsCurrent:
			prefix = " "
		case !noColorFlag && !profile.IsCurrent:
			prefix = normalColor
		}
		println(prefix + profile.Name)
	}
	return nil
}

func printVersion(c *cli.Context) error {
	fmt.Printf("Version %s", c.App.Version)
	return nil
}

func printProfile(c *cli.Context) error  {
	aws, err := initAwsctx()
	if err != nil {
		return err
	}
	profile, err := getCurrentProfile(aws)
	if err != nil {
		return err
	}
	fmt.Println(profile)
	return nil
}

func getCurrentProfile(aws *awsctx.Awsctx) (string, error) {
	profiles, err := aws.GetProfiles()
	if err != nil {
		return "", err
	}
	for _, p := range profiles {
		if p.IsCurrent {
			return p.Name, nil
		}
	}
	return "", fmt.Errorf("could not find current profile")
}
