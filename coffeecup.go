package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jxskiss/mcli"
	"github.com/ttacon/chalk"
	"golang.org/x/term"
)

func main() {
	mcli.Add("login", LoginCommand, "Login to CoffeeCup")
	mcli.Add("start", StartCommand, "Starts/Resumes a time entry. Needs a project alias as argument. Optionally, you can provide a comment that will be appeneded to any existing comment.\n\nExamples:\n - coffeecup start myproject\n - coffeecup start myproject \"This is a comment\"")
	mcli.Add("stop", StopCommand, "Stops any running time entries")
	mcli.Add("today", TodayCommand, "Lists today's time entries")

	mcli.AddGroup("projects", "Lists projects and assign aliases to your active projects")
	mcli.Add("projects list", ProjectsListCommand, "Lists all active projects")
	mcli.Add("projects alias", ProjectAliasCommand, "Lists the known aliases or sets new ones. Use \"coffeecup projects list\" to figure out the ID of your project.\n\nExamples:\n - coffeecup projects alias\n - coffeecup projects alias 90454 myproject")

	// Enable shell auto-completion, see `program completion -h` for help.
	// mcli.AddCompletion()
	mcli.AddHelp()

	mcli.Run()
}

func LoginCommand() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter company prefix (the \"acme\" in \"acme.coffeecup.app\"): ")
	companyName, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	fmt.Print("Enter Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Println()

	accessToken, refreshToken, err := LoginWithPassword(strings.TrimSpace(companyName), strings.TrimSpace(username), string(bytePassword))
	if err != nil {
		panic(err)
	}

	StoreTokens(accessToken, refreshToken)

	userId, err := GetUserId()
	if err != nil {
		panic(err)
	}

	StoreUserId(userId)
	fmt.Printf("Successfully logged in to %s as %s\n", strings.TrimSpace(companyName), strings.TrimSpace(username))
}

func LoginUsingRefreshToken() error {
	accessToken, refreshToken, err := LoginWithRefreshToken()
	if err != nil {
		return err
	}

	return StoreTokens(accessToken, refreshToken)
}

func ProjectsListCommand() {
	projects, err := GetProjects()

	// retry if unauthorized
	if err != nil && err.Error() == "unauthorized" {
		err = LoginUsingRefreshToken()
		if err != nil {
			fmt.Printf("Please login first using %s'coffeecup login'%s\n", chalk.Cyan, chalk.Reset)
			os.Exit(1)
		}

		projects, err = GetProjects()
	}

	if err != nil {
		panic(err)
	}

	for _, project := range projects {
		fmt.Printf("%-8d %s\n", project.Id, project.Name)
	}
}

func ProjectAliasCommand() {
	var args struct {
		ProjectId string `cli:"id, The ID of the project (optional)"`
		Alias     string `cli:"alias, The alias of the project (optional)"`
	}
	_, err := mcli.Parse(&args)
	if err != nil {
		panic(err)
	}

	cfg, _ := ReadConfig()
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

	if cfg.Projects == nil {
		cfg.Projects = make(map[string]ProjectConfig)
	}

	project, ok := cfg.Projects[args.Alias]
	if !ok {
		project = ProjectConfig{}
	}

	lastTimeEntryForProject, err := GetLastTimeEntryForProject(project.Id)
	// retry if unauthorized
	if err != nil && err.Error() == "unauthorized" {
		err = LoginUsingRefreshToken()
		if err != nil {
			fmt.Printf("Please login first using %s'coffeecup login'%s\n", chalk.Cyan, chalk.Reset)
			os.Exit(1)
		}
		lastTimeEntryForProject, err = GetLastTimeEntryForProject(project.Id)
	}

	if err != nil {
		fmt.Printf("%sCouldn't determine default Task ID for this project. Please run this command again or configurate it manually in your coffeecup.toml.%s\n", chalk.Red, chalk.Reset)
	} else {
		project.DefaultTaskId = lastTimeEntryForProject.TaskId
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
		err = LoginUsingRefreshToken()
		if err != nil {
			fmt.Printf("Please login first using %s'coffeecup login'%s\n", chalk.Cyan, chalk.Reset)
			os.Exit(1)
		}
		timeEntries, err = GetTodaysTimeEntries()
	}

	if err != nil {
		panic(err)
	}

	cfg, _ := ReadConfig()
	projectConfigs := cfg.Projects
	var targetedProjectId int
	for _, project := range projectConfigs {
		if project.Alias == args.Alias {
			targetedProjectId = project.Id
			break
		}
	}
	if targetedProjectId == 0 {
		fmt.Printf("Project alias %s'%s'%s not found 😱\nRun %s'coffeecup help projects alias'%s to learn how to set an alias.\n", chalk.Red, args.Alias, chalk.Reset, chalk.Cyan, chalk.Reset)
		os.Exit(1)
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
			ProjectId: projectId,
			Day:       today,
			Duration:  0,
			Sorting:   len(timeEntries) + 1,
			Running:   true,
			Comment:   comment,
			// get the actual task ID for this project from /taskAssignments?where={ "project": 90545 }
			// and/or store them as the DefaultTaskId in the config
			// To get the default: get the last timeEntry for this project and use that task id?
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
		err = LoginUsingRefreshToken()
		if err != nil {
			fmt.Printf("Please login first using %s'coffeecup login'%s\n", chalk.Cyan, chalk.Reset)
			os.Exit(1)
		}
		timeEntries, err = GetTodaysTimeEntries()
	}

	if err != nil {
		panic(err)
	}

	cfg, _ := ReadConfig()
	projectConfigs := cfg.Projects
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
		err = LoginUsingRefreshToken()
		if err != nil {
			fmt.Printf("Please login first using %s'coffeecup login'%s\n", chalk.Cyan, chalk.Reset)
			os.Exit(1)
		}
		timeEntries, err = GetTodaysTimeEntries()
	}

	if err != nil {
		panic(err)
	}

	cfg, _ := ReadConfig()
	projectConfigs := cfg.Projects

	var projects []Project

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

		// if we don't have an alias for this project, use the full name
		if projectAlias == "" {
			if projects == nil {
				projects, err = GetProjects()
				if err != nil {
					panic(err)
				}
			}
			for _, project := range projects {
				if project.Id == timeEntry.ProjectId {
					projectAlias = project.Name
					break
				}
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
