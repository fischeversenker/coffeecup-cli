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
	mcli.Add("start", StartCommand, "Starts/Resumes a time entry. Needs a project alias as argument. Optionally, you can provide a comment that will be appeneded to any existing comment.")
	mcli.Add("stop", StopCommand, "Stops any running time entries")
	mcli.Add("today", TodayCommand, "Lists today's time entries")
	mcli.AddAlias("status", "today")

	mcli.Add("version", func() { fmt.Println("v0.0.12") }, "Prints the version of CoffeeCup CLI")

	mcli.AddGroup("projects", "Lists projects and assign aliases to your active projects")
	mcli.Add("projects list", ProjectsListCommand, "Lists all active projects")
	mcli.Add("projects alias", ProjectAliasCommand, "Lists the known aliases or sets new ones. Use the \"projects list\" command to figure out the ID of your project.")

	// Enable shell auto-completion, see `program completion -h` for help.
	// mcli.AddCompletion()
	mcli.AddHelp()

	mcli.Run()
}

func LoginCommand() {
	reader := bufio.NewReader(os.Stdin)
	loggedInUser, err := GetUser()
	if err == nil {
		fmt.Printf("You are already logged in as \"%s\".\n", loggedInUser.Email)
		fmt.Print("Do you want to login as someone else? (y/n) ")
		answer, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		if strings.TrimSpace(answer) == "n" {
			return
		}
	}

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

	user, err := GetUser()
	if err != nil {
		panic(err)
	}

	StoreUserId(user.Id)
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
			fmt.Println(chalk.Yellow.Color("Please login first using the 'login' command"))
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

	// add to config
	cfg, _ := ReadConfig()
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]ProjectConfig)
	}

	for _, project := range projects {
		currentConfig := cfg.Projects[strconv.Itoa(project.Id)]
		cfg.Projects[strconv.Itoa(project.Id)] = ProjectConfig{
			Id:            project.Id,
			Name:          project.Name,
			Alias:         currentConfig.Alias,
			DefaultTaskId: currentConfig.DefaultTaskId,
		}
	}

	WriteConfig(cfg)
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
		for _, project := range cfg.Projects {
			if project.Alias != "" {
				fmt.Printf("%-10s %-20s (ID: %d)\n", project.Alias, project.Name, project.Id)
			}
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

	project, ok := cfg.Projects[args.ProjectId]
	if !ok {
		project = ProjectConfig{}
	}

	lastTimeEntryForProject, err := GetLastTimeEntryForProject(project.Id)
	// retry if unauthorized
	if err != nil && err.Error() == "unauthorized" {
		err = LoginUsingRefreshToken()
		if err != nil {
			fmt.Println(chalk.Yellow.Color("Please login first using the 'login' command"))
			os.Exit(1)
		}
		lastTimeEntryForProject, err = GetLastTimeEntryForProject(project.Id)
	}

	if err != nil {
		fmt.Printf("%sCouldn't determine your default Task ID for project '%s'%s.\n", chalk.Red, project.Name, chalk.Reset)
		fmt.Printf("This will prevent this program from properly starting a time entry for this project.\n")
		fmt.Printf("You probably haven't yet booked time on this project. If you have, please run this command again.\n")
		fmt.Printf("If you are adventurous, please configure the default task ID for this project (ID: %d) manually in %s%s%s.\n", project.Id, chalk.Cyan, GetConfigPath(), chalk.Reset)
	} else {
		project.DefaultTaskId = lastTimeEntryForProject.TaskId
	}

	project.Id, _ = strconv.Atoi(args.ProjectId)
	project.Alias = args.Alias
	cfg.Projects[args.ProjectId] = project
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
			fmt.Println(chalk.Yellow.Color("Please login first using the 'login' command"))
			os.Exit(1)
		}
		timeEntries, err = GetTodaysTimeEntries()
	}

	if err != nil {
		panic(err)
	}

	cfg, _ := ReadConfig()
	projectConfigs := cfg.Projects
	var targetedProject ProjectConfig
	for _, project := range projectConfigs {
		if project.Alias == args.Alias {
			targetedProject = project
			break
		}
	}
	if targetedProject.Id == 0 {
		fmt.Printf("Project alias %s'%s'%s not found 😱\nRun the %s'help projects alias'%s command to learn how to set an alias.\n", chalk.Red, args.Alias, chalk.Reset, chalk.Cyan, chalk.Reset)
		os.Exit(1)
	}

	resumedExistingTimeEntry := false
	wasRunningAlready := false
	for _, timeEntry := range timeEntries {
		if timeEntry.Running {
			if targetedProject.Id == timeEntry.ProjectId {
				fmt.Printf("%s%s%s is running already\n", chalk.Green, targetedProject.Alias, chalk.Reset)
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
				err := UpdateTimeEntry(timeEntry)
				if err != nil {
					panic(err)
				}
			}
		} else {
			if targetedProject.Id == timeEntry.ProjectId {
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
				fmt.Printf("Resumed existing time entry for %s%s%s\n", chalk.Green, args.Alias, chalk.Reset)
				resumedExistingTimeEntry = true
			}
		}
	}

	if !resumedExistingTimeEntry && !wasRunningAlready {
		// start a new time entry
		projectId := targetedProject.Id
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
			TaskId:       targetedProject.DefaultTaskId,
			TrackingType: "WORK",
			UserId:       GetUserIdFromConfig(),
		})
		if err != nil {
			fmt.Println("Error creating new time entry:")

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
			fmt.Println(chalk.Yellow.Color("Please login first using the 'login' command"))
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
	var args struct {
		Color bool `cli:"-c, --color, enable colors in the output"`
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
			fmt.Println(chalk.Yellow.Color("Please login first using the 'login' command"))
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

	var overallTime int
	var longestComment int

	for _, timeEntry := range timeEntries {
		overallTime += timeEntry.Duration
		longestComment = max(longestComment, len(timeEntry.Comment))
	}

	// todo: use more colors with chalk
	for _, timeEntry := range timeEntries {
		hours := timeEntry.Duration / 3600
		minutes := (timeEntry.Duration % 3600) / 60
		timeString := fmt.Sprintf("%02dh %02dm", hours, minutes)

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
		if timeEntry.Running {
			fmt.Printf("%-10s | ⌛ %s | 📝 %-*s\n", projectAlias, timeString, longestComment, comment)
		} else {
			if args.Color {
				fmt.Printf("%-19s %s    %s %s 📝 %-*s\n", chalk.Dim.TextStyle(projectAlias), chalk.Dim.TextStyle("|"), chalk.Dim.TextStyle(timeString), chalk.Dim.TextStyle("|"), longestComment, chalk.Dim.TextStyle(comment))
			} else {
				fmt.Printf("%-10s |    %s | 📝 %-*s\n", projectAlias, timeString, longestComment, comment)
			}
		}
	}

	hours := overallTime / 3600
	minutes := (overallTime % 3600) / 60
	time := fmt.Sprintf("%02dh %02dm", hours, minutes)
	if args.Color {
		fmt.Printf("%-19s %s    %s\n", chalk.Dim.TextStyle("total"), chalk.Dim.TextStyle("|"), chalk.Dim.TextStyle(time))
	} else {
		fmt.Printf("%-10s |    %s\n", "total", time)
	}
}
