package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"

	"code.tobolaski.com/brendan/ejrnl"
	"code.tobolaski.com/brendan/ejrnl/storage"
	"code.tobolaski.com/brendan/ejrnl/workflows"
)

var version = "0.0.1"

var tempFlag = cli.StringFlag{
	Name:  "temp-dir",
	Usage: "Specifies the temporary directory to write the temporary files to.",
	Value: os.TempDir(),
}

func main() {
	app := cli.NewApp()

	app.Name = "ejrnl"
	app.Usage = "An encrypted journal application"
	app.Version = version
	app.Authors = []cli.Author{cli.Author{
		Name:  "Brendan Tobolaski",
		Email: "brendan@tobolaski.com",
	}}

	var configPath string
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Usage: "Specifies the config file to use",
			Value: "~/.config/ejrnl/ejrnl.yml",
		},
	}
	app.Before = func(c *cli.Context) error {
		user, err := user.Current()
		if err != nil {
			return err
		}
		configPath = strings.Replace(c.String("config"), "~", user.HomeDir, -1)
		return nil
	}
	app.Commands = []cli.Command{
		{
			Name:  "init",
			Usage: "Creates a new journal",
			Flags: []cli.Flag{
				cli.UintFlag{
					Name:  "pow",
					Usage: "Configures the workfactor for scrypt",
					Value: 19,
				},
				cli.StringFlag{
					Name:  "destination",
					Usage: "Configures where the journal is stored",
					Value: "~/journal",
				},
			},
			Action: func(c *cli.Context) error {
				if err := os.MkdirAll(path.Dir(configPath), 0750); err != nil {
					return err
				}
				if _, err := os.Stat(configPath); !os.IsNotExist(err) {
					return errors.New("Configuration directory already exists")
				}
				config := workflows.DefaultConfig()
				config.StorageDirectory = c.String("destination")
				config.Pow = c.Uint("pow")
				data, err := yaml.Marshal(config)
				if err != nil {
					return err
				}
				err = ioutil.WriteFile(configPath, data, 0600)
				if err != nil {
					return err
				}
				password, err := getPassword()
				if err != nil {
					return err
				}

				driver, err := storage.NewDriver(config, password)
				password = ""
				if _, ok := err.(*storage.NeedsInit); !ok {
					return err
				}
				return workflows.Init(driver)
			},
		},
		{
			Name:  "import",
			Usage: "Adds the specified file into the journal",
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 1 {
					return errors.New("import requires 1 argument which is the file to import")
				}
				driver, err := standardLoad(configPath)
				if err != nil {
					return err
				}

				return workflows.Import(c.Args()[0], driver)
			},
		},
		{
			Name:  "print",
			Usage: "Prints out the most recent entries",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "count",
					Usage: "The number of entries to output. If it is 0 or less, all entries are output",
					Value: 0,
				},
			},
			Action: func(c *cli.Context) error {
				driver, err := standardLoad(configPath)
				if err != nil {
					return err
				}

				return workflows.Print(driver, c.Int("count"))
			},
		},
		{
			Name:  "list",
			Usage: "Lists the ids and dates of the most recent entries",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "count",
					Value: 0,
					Usage: "The number of results to return. If it is <= 0, it returns all of the entries",
				},
			},
			Action: func(c *cli.Context) error {
				driver, err := standardLoad(configPath)
				if err != nil {
					return err
				}

				return workflows.ListEntries(driver, c.Int("count"))
			},
		},
		{
			Name:  "new",
			Usage: "Creates a new entry",
			Flags: []cli.Flag{tempFlag},
			Action: func(c *cli.Context) error {
				driver, err := standardLoad(configPath)
				if err != nil {
					return err
				}
				return workflows.NewEntry(driver, c.String("temp-dir"))
			},
		},
		{
			Name:  "edit",
			Usage: "edits an existing journal entry. Takes an id as an argument.",
			Flags: []cli.Flag{tempFlag},
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 1 {
					return errors.New("edit takes 1 argument which is an entry's id")
				}
				driver, err := standardLoad(configPath)
				if err != nil {
					return err
				}
				return workflows.EditEntry(driver, c.Args()[0], c.String("temp-dir"))
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Printf("Failed to complete because %s", err)
		os.Exit(1)
	}
}

func getPassword() (string, error) {
	fmt.Printf("Password: ")
	raw, err := gopass.GetPasswd()
	return string(raw), err
}

func readConfig(path string) (ejrnl.Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ejrnl.Config{}, err
	}
	entry := &ejrnl.Config{}
	err = yaml.Unmarshal(data, entry)
	return *entry, err
}

func standardLoad(configPath string) (*storage.Driver, error) {
	config, err := readConfig(configPath)
	if err != nil {
		return &storage.Driver{}, err
	}
	password, err := getPassword()
	if err != nil {
		return &storage.Driver{}, err
	}
	return storage.NewDriver(config, password)
}
