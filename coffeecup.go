package main

import (
	"fmt"
	"strconv"

	mcli "github.com/jxskiss/mcli"
	"github.com/ttacon/chalk"
)

func main() {
	mcli.Add("login", LoginCommand, "Login to CoffeeCup")
	mcli.Add("start", StartCommand, "Start/Resume time entry")
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
		ProjectId int    `cli:"id, The ID of the project"`
		Alias     string `cli:"alias, The alias of the project"`
	}
	_, err := mcli.Parse(&args)
	if err != nil {
		panic(err)
	}

	cfg := ReadConfig()
	if cfg.Projects.Aliases == nil {
		cfg.Projects.Aliases = make(map[string]string)
	}

	if (args.ProjectId != 0) && (args.Alias == "") {
		fmt.Println("Please provide an alias for the project")
		return
	} else if (args.ProjectId == 0) && (args.Alias == "") {
		fmt.Println("Configured aliases:")
		for projectId, alias := range cfg.Projects.Aliases {
			fmt.Printf("- %s: %s\n", projectId, alias)
		}
		return
	} else {
		cfg.Projects.Aliases[strconv.Itoa(args.ProjectId)] = args.Alias
		WriteConfig(cfg)
	}
}

func StartCommand() {
	var args struct {
		Alias string `cli:"alias, The alias of the project"`
	}
	_, err := mcli.Parse(&args)
	if err != nil {
		panic(err)
	}

	if args.Alias == "" {
		fmt.Println("Please provide a project alias")
		// todo: use previously used project for convenience
		return
	} else {
		timeEntries, _ := GetTodaysTimeEntries()
		projectAliases := ReadConfig().Projects.Aliases
		resumedExisting := false
		for _, timeEntry := range timeEntries {
			if projectAliases[strconv.Itoa(timeEntry.Project)] == args.Alias {
				if timeEntry.Running {
					fmt.Printf("%s%s%s is running already\n", chalk.Green, args.Alias, chalk.Reset)
					return
				}

				// resume the time entry
				// ResumeTimeEntry(timeEntry.Id)
				resumedExisting = true
				return
			} else {
				// if it's running, stop it
				// StopTimeEntry(timeEntry.Id)
			}
		}

		if !resumedExisting {
			// start a new time entry
			// CreateNewRunningTimeEntry(projectId)
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

	cfg := ReadConfig()
	aliases := cfg.Projects.Aliases

	if timeEntries == nil {
		fmt.Println("No time entries for today")
		return
	}

	for _, timeEntry := range timeEntries {
		hours := timeEntry.Duration / 3600
		minutes := (timeEntry.Duration % 3600) / 60

		// todo: use more colors with chalk
		fmt.Printf("Project: %s\nDuration: %d:%d\nComment:\n%s\n\n", aliases[strconv.Itoa(timeEntry.Project)], hours, minutes, timeEntry.Comment)
	}
}
