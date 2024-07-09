package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	mcli "github.com/jxskiss/mcli"
	"github.com/ttacon/chalk"
)

func main() {
	mcli.Add("login", LoginCommand, "Login to CoffeeCup")
	mcli.Add("start", StartCommand, "Start/Resume time entry")
	mcli.Add("stop", StopCommand, "Stop any running time entries")
	mcli.Add("today", TodayCommand, "Show today's time entries")

	mcli.Add("projects list", ProjectsListCommand, "Lists all projects")
	mcli.Add("projects alias", ProjectAliasCommand, "Lists the known aliases or sets new ones")

	// Enable shell auto-completion, see `program completion -h` for help.
	mcli.AddCompletion()

	mcli.Run()
}

func LoginCommand() {
	var args struct {
		CompanyUrl string `cli:"#R, -c, --company, The prefix of the company's CoffeeCup instance (the \"amazon\" in \"amazon.coffeecup.app\")"`
		Username   string `cli:"#R, -u, --username, The username of the user"`
		Password   string `cli:"#R, -p, --password, The password of the user"`
	}
	_, err := mcli.Parse(&args)
	if err != nil {
		panic(err)
	}

	accessToken, refreshToken, err := LoginWithPassword(args.CompanyUrl, args.Username, args.Password)
	if err != nil {
		panic(err)
	}

	StoreTokens(accessToken, refreshToken)

	userId, err := GetUserId()
	if err != nil {
		panic(err)
	}

	StoreUserId(userId)
	fmt.Printf("Successfully logged in as %s\n", args.Username)
}

func LoginUsingRefreshToken() {
	accessToken, refreshToken, err := LoginWithRefreshToken()
	if err != nil {
		panic(err)
	}

	StoreTokens(accessToken, refreshToken)
}

func ProjectsListCommand() {
	projects, err := GetProjects()

	// retry if unauthorized
	if err != nil && err.Error() == "unauthorized" {
		LoginUsingRefreshToken()
		projects, err = GetProjects()
	}

	if err != nil {
		panic(err)
	}

	for _, project := range projects {
		fmt.Printf("%d: %s\n", project.Id, project.Name)
	}
}

func ProjectAliasCommand() {
	var args struct {
		ProjectId string `cli:"id, The ID of the project"`
		Alias     string `cli:"alias, The alias of the project"`
	}
	_, err := mcli.Parse(&args)
	if err != nil {
		panic(err)
	}

	cfg := ReadConfig()
	if (args.ProjectId == "") && (args.Alias == "") {
		fmt.Println("Configured aliases:")
		for _, project := range cfg.Projects {
			fmt.Printf("%d: %s\n", project.Id, project.Alias)
		}
		return
	}

	if (args.ProjectId != "") && (args.Alias == "") {
		fmt.Println("Please provide an alias for the project")
		os.Exit(1)
	}

	project, ok := cfg.Projects[args.Alias]
	if !ok {
		project = ProjectConfig{}
	}

	project.Id, _ = strconv.Atoi(args.ProjectId)
	project.Alias = args.Alias
	cfg.Projects[args.Alias] = project
	WriteConfig(cfg)
}

func StartCommand() {
	var args struct {
		Alias   string `cli:"#R, alias, The alias of the project"`
		Comment string `cli:"comment, The comment for the time entry"`
	}
	_, err := mcli.Parse(&args)
	if err != nil {
		panic(err)
	}

	timeEntries, err := GetTodaysTimeEntries()
	// retry if unauthorized
	if err != nil && err.Error() == "unauthorized" {
		LoginUsingRefreshToken()
		timeEntries, err = GetTodaysTimeEntries()
	}

	if err != nil {
		panic(err)
	}

	projectConfigs := ReadConfig().Projects
	var targetedProjectId int
	for _, project := range projectConfigs {
		if project.Alias == args.Alias {
			targetedProjectId = project.Id
			break
		}
	}
	resumedExistingTimeEntry := false
	wasRunningAlready := false
	for _, timeEntry := range timeEntries {
		if timeEntry.Running {
			if targetedProjectId == timeEntry.ProjectId {
				fmt.Printf("%s%s%s is running already\n", chalk.Green, args.Alias, chalk.Reset)
				if args.Comment != "" {
					if timeEntry.Comment == "" {
						timeEntry.Comment = "- " + args.Comment
					} else {
						timeEntry.Comment = timeEntry.Comment + "\n- " + args.Comment
					}
					err := UpdateTimeEntry(timeEntry)
					if err != nil {
						panic(err)
					}
					fmt.Printf("Added comment '%s'\n", args.Comment)
				}
				wasRunningAlready = true
			} else {
				// wrong project is running, stop it
				timeEntry.Running = false
				var projectAlias string
				for _, project := range projectConfigs {
					if project.Id == timeEntry.ProjectId {
						projectAlias = project.Alias
						break
					}
				}
				err := UpdateTimeEntry(timeEntry)
				if err != nil {
					panic(err)
				}
				fmt.Printf("Stopped %s%s%s\n", chalk.Red, projectAlias, chalk.Reset)
			}
		} else {
			if targetedProjectId == timeEntry.ProjectId {
				// not running, resume it
				timeEntry.Running = true
				if args.Comment != "" {
					if timeEntry.Comment == "" {
						timeEntry.Comment = "- " + args.Comment
					} else {
						timeEntry.Comment = timeEntry.Comment + "\n- " + args.Comment
					}
				}
				err := UpdateTimeEntry(timeEntry)
				if err != nil {
					panic(err)
				}
				fmt.Printf("Resumed previous time entry for %s%s%s\n", chalk.Green, args.Alias, chalk.Reset)
				resumedExistingTimeEntry = true
			}
		}
	}

	if !resumedExistingTimeEntry && !wasRunningAlready {
		// start a new time entry
		projectId := targetedProjectId
		today := time.Now().Format("2006-01-02")
		var comment string
		if args.Comment != "" {
			comment = "- " + args.Comment
		}
		err := CreateTimeEntry(NewTimeEntry{
			ProjectId:    projectId,
			Day:          today,
			Duration:     0,
			Sorting:      len(timeEntries) + 1,
			Running:      true,
			Comment:      comment,
			TaskId:       1095, // hardcoded task id for "Frontend" for now
			TrackingType: "WORK",
			UserId:       GetUserIdFromConfig(),
		})
		if err != nil {
			fmt.Printf("%s%s%s\n", chalk.Red, err, chalk.Reset)
			os.Exit(1)
		}

		fmt.Printf("Started new time entry for %s%s%s\n", chalk.Green, args.Alias, chalk.Reset)
	}
}

func StopCommand() {
	timeEntries, err := GetTodaysTimeEntries()
	// retry if unauthorized
	if err != nil && err.Error() == "unauthorized" {
		LoginUsingRefreshToken()
		timeEntries, err = GetTodaysTimeEntries()
	}

	if err != nil {
		panic(err)
	}

	projectConfigs := ReadConfig().Projects
	for _, timeEntry := range timeEntries {
		if timeEntry.Running {
			timeEntry.Running = false
			var projectAlias string
			for _, project := range projectConfigs {
				if project.Id == timeEntry.ProjectId {
					projectAlias = project.Alias
					break
				}
			}
			fmt.Printf("Stopped %s%s%s\n", chalk.Red, projectAlias, chalk.Reset)
			err := UpdateTimeEntry(timeEntry)
			if err != nil {
				panic(err)
			}
		}
	}
}

func TodayCommand() {
	timeEntries, err := GetTodaysTimeEntries()

	// retry if unauthorized
	if err != nil && err.Error() == "unauthorized" {
		LoginUsingRefreshToken()
		timeEntries, err = GetTodaysTimeEntries()
	}

	if err != nil {
		panic(err)
	}

	projectConfigs := ReadConfig().Projects

	if len(timeEntries) == 0 {
		fmt.Println("No time entries for today")
		return
	}

	// todo: use more colors with chalk
	for _, timeEntry := range timeEntries {
		hours := timeEntry.Duration / 3600
		minutes := (timeEntry.Duration % 3600) / 60

		var projectAlias string
		for _, project := range projectConfigs {
			if project.Id == timeEntry.ProjectId {
				projectAlias = project.Alias
				break
			}
		}

		comment := strings.ReplaceAll(timeEntry.Comment, "\n", " ")
		var isRunning string
		if timeEntry.Running {
			isRunning = "⌛"
		} else {
			isRunning = "  "
		}
		fmt.Printf("%-10s | %s %02dh %02dm | 📝 %-10s\n", projectAlias, isRunning, hours, minutes, comment)
	}
}
