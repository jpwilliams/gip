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
	"github.com/gosuri/uitable"
	"github.com/jinzhu/configor"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
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

type Log struct {
	Timestamp string
	Time      string
	Repo      string
	Message   string
	Author    string
	Details   string
	Sign      string
}

type byTimestamp []Log

func (c byTimestamp) Len() int           { return len(c) }
func (c byTimestamp) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c byTimestamp) Less(i, j int) bool { return c[i].Timestamp <= c[j].Timestamp }

var config Config

func main() {
	configPath := "/Users/jackwilliams/.gip.json"
	configor.Load(&config, configPath)

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

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "fetch",
			Usage: "Fetch remotes from repositories before displaying",
		},
	}

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
					Name:   "list",
					Usage:  "List groups",
					Action: listGroups,
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
				cli.IntFlag{
					Name:  "n",
					Usage: "Maximum number of logs to return",
					Value: 0,
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

func getLog(bar *uiprogress.Bar, name string, path string, after string, before string, waiter *sync.WaitGroup, ret chan Log) {
	defer waiter.Done()
	defer bar.Incr()

	cmdStr := []string{
		"--no-pager",
		"log",
		"--all",
		"--date=format:%a %R",
		"--pretty=%ct ||| %cd ||| %s ||| %an ||| %G? ||| %d",
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
		split := strings.Split(line, " ||| ")

		if len(split) < 4 {
			continue
		}

		parsed := Log{
			Timestamp: split[0],
			Time:      split[1],
			Repo:      name,
			Message:   split[2],
			Author:    split[3],
			Sign:      split[4],
		}

		if len(split) == 6 {
			parsed.Details = split[5]
		}

		ret <- parsed
	}
}

func view(c *cli.Context) error {
	if c.Args().First() == "" {
		cli.ShowSubcommandHelp(c)

		return nil
	}

	viewLogs(c.Args().First(), false, c.String("after"), c.String("before"), c.Int("n"))

	return nil
}

func viewLogs(repo string, clear bool, after string, before string, max int) {
	repos := config.Groups[repo].Repos
	retChan := make(chan Log, 500)
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
	var logs []Log

	for line := range retChan {
		logs = append(logs, line)
	}

	sort.Sort(byTimestamp(logs))

	if max > 0 {
		logs = logs[max:]
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Time", "Repo", "Author", "Commit"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("|")
	termWidth, _ := terminal.Width()
	table.SetColWidth(int(termWidth))

	for _, log := range logs {
		table.Append([]string{log.Time, log.Repo, log.Author + " (" + log.Sign + ")", log.Message})
	}

	table.Render()
}
	}
}
