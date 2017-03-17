package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gosuri/uiprogress"
	"github.com/jinzhu/configor"
	"github.com/urfave/cli"
)

type Config struct {
	Groups map[string]Group
	Repos  map[string]Repo
}

type Group struct {
	Repos []string
}

type Repo struct {
	Path   string
	Groups []string
}

var config Config

func main() {
	configor.Load(&config, "/Users/jackwilliams/.gip.json")

	app := cli.NewApp()
	app.Version = "0.0.0"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Jack Williams",
			Email: "jack@wildfire.gg",
		},
	}
	app.Name = "gip"
	app.Usage = "List git commits across grouped repositories within a given time period"

	app.Commands = []cli.Command{
		{
			Name:    "group",
			Aliases: []string{"g"},
			Usage:   "Manage repository groups",
			Subcommands: []cli.Command{
				{
					Name:      "create",
					Usage:     "Create a new group",
					ArgsUsage: "[name]",
					Action: func(c *cli.Context) error {
						return nil
					},
				}, {
					Name:  "list",
					Usage: "List groups",
					Action: func(c *cli.Context) error {
						return nil
					},
				}, {
					Name:      "add",
					Usage:     "Add a repository to a group",
					ArgsUsage: "[repo] [group]",
					Action: func(c *cli.Context) error {
						return nil
					},
				}, {
					Name:      "rename",
					Usage:     "Rename a group",
					ArgsUsage: "[group] [name]",
					Action: func(c *cli.Context) error {
						return nil
					},
				}, {
					Name:      "remove",
					Usage:     "Delete a group",
					ArgsUsage: "[group]",
					Action: func(c *cli.Context) error {
						return nil
					},
				},
			},
		}, {
			Name:    "repo",
			Aliases: []string{"r"},
			Usage:   "Manage repositories",
		}, {
			Name:      "view",
			Aliases:   []string{"v, vg"},
			Usage:     "List commits within a group",
			ArgsUsage: "[group]",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "watch, w",
					Usage: "Persistently refresh the results to get a live feed of updates",
				},
				cli.StringFlag{
					Name:  "after, a",
					Value: "12am",
					Usage: "Specify a 'git log'-compatible time period to list commits since",
				},
				cli.StringFlag{
					Name:  "before, b",
					Usage: "Specify a 'git log'-compatible time period to list commits until",
				},
			},
			Action: view,
		}, {
			Name:    "viewrepo",
			Aliases: []string{"vr"},
			Usage:   "List commits within a repository",
		},
	}

	app.Run(os.Args)
}

func getLog(bar *uiprogress.Bar, name string, path string, after string, before string, waiter *sync.WaitGroup, ret chan string) {
	defer waiter.Done()
	defer bar.Incr()

	cmdStr := []string{
		"--no-pager",
		"log",
		"--all",
		"--date=format:%a %R",
		"--pretty=%ct %cd %CblueNAME%Creset%Cgreen %s%Creset %Cred%d%Creset - %an",
		"--reverse",
		"--after=" + after,
	}

	if before != "" {
		cmdStr = append(cmdStr, "--before="+before)
	}

	cmd := exec.Command("git", cmdStr...)
	cmd.Dir = path
	emojify := exec.Command("emojify")

	pipe, _ := cmd.StdoutPipe()
	defer pipe.Close()
	emojify.Stdin = pipe
	cmd.Start()
	out, _ := emojify.Output()
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	if len(lines) == 0 || lines[0] == "" {
		return
	}

	for _, line := range lines {
		parsed := strings.Replace(line, "NAME", name+" -", 1)
		ret <- parsed
	}
}

func view(c *cli.Context) error {
	if c.Args().First() == "" {
		cli.ShowSubcommandHelp(c)

		return nil
	}

	viewLogs(c.Args().First(), false, c.String("after"), c.String("before"))

	return nil
}

func viewLogs(repo string, clear bool, after string, before string) {
	repos := config.Groups[repo].Repos
	retChan := make(chan string, 500)
	waiter := &sync.WaitGroup{}

	uiprogress.Start()
	bar := uiprogress.AddBar(len(repos))
	bar.AppendCompleted()
	bar.PrependElapsed()

	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return "(" + strconv.Itoa(b.Current()) + "/" + strconv.Itoa(len(repos)) + ")"
	})

	for _, repo := range repos {
		waiter.Add(1)
		go getLog(bar, repo, config.Repos[repo].Path, after, before, waiter, retChan)
	}

	waiter.Wait()
	close(retChan)
	uiprogress.Stop()
	var logs []string

	for line := range retChan {
		logs = append(logs, line)
	}

	sort.Strings(logs)

	if clear {
		clr := exec.Command("clear")
		clr.Stdout = os.Stdout
		clr.Run()
		clr.Wait()
		fmt.Println("cleared")
	}

	for _, line := range logs {
		fmt.Println(line[11:])
	}
}
