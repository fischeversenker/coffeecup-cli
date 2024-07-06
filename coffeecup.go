package main

import (
	"fmt"

	mcli "github.com/jxskiss/mcli"
)

func main() {
	mcli.Add("login", loginCommand, "Login to CoffeeCup")
	mcli.Add("start", startCommand, "Start/Resume time entry")

	mcli.AddGroup("projects", "This is a command group called cmd2")
	mcli.Add("projects list", projectsListCommand, "Do something with cmd2 sub1")
	mcli.Add("projects alias", projectAliasCommand, "Do something with cmd2 sub1")

	// Enable shell auto-completion, see `program completion -h` for help.
	mcli.AddCompletion()

	mcli.Run()
}

func loginCommand() {
	var args struct {
		CompanyUrl string `cli:"#R, -c, --company, The prefix of the company's CoffeeCup instance (the \"amazon\" in \"amazon.coffeecup.app\")"`
		Username   string `cli:"#R, -u, --username, The username of the user"`
		Password   string `cli:"#R, -p, --password, The password of the user"`
	}
	_, err := mcli.Parse(&args)
	if err != nil {
		panic(err)
	}

	accesstoken, responsetoken, err := login(args.CompanyUrl, args.Username, args.Password)
	if err != nil {
		panic(err)
	}

	storeToken(accesstoken, responsetoken)
	fmt.Printf("Successfully logged in as %s\n", args.Username)
}

func projectsListCommand() {
	projects, err := getProjects()
	if err != nil {
		panic(err)
	}

	for _, project := range projects {
		fmt.Printf("%d: %s\n", project.Id, project.Name)
	}
}

func projectAliasCommand() {
	var args struct {
		ProjectId int    `cli:"#R, id, The ID of the project"`
		Alias     string `cli:"#R, alias, The alias of the project"`
	}
	_, err := mcli.Parse(&args)
	if err != nil {
		panic(err)
	}

	cfg := readConfig()
	if cfg.Projects.Aliases == nil {
		cfg.Projects.Aliases = make(map[string]string)
	}

	cfg.Projects.Aliases[args.Alias] = fmt.Sprint(args.ProjectId)
	writeConfig(cfg)
}

func startCommand() {
	var args struct {
		Alias string `cli:"alias, The alias of the project"`
	}
	_, err := mcli.Parse(&args)
	if err != nil {
		panic(err)
	}

	if args.Alias == "" {
		fmt.Println("Please provide a project alias")
		return
	} else {
		fmt.Println("Checking if there is a time entry from today that I can resume")
		// get existing time entries
		// find the most recent one from today
		// or find the one that is currently running
		// if there is a running one, don't do anything
		// if there is one from today, resume it
		// if there is none, fail and request an alias
	}
}
