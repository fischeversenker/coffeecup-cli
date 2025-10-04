# Aerion CLI

> [!NOTE]
> This is a private project! I'm not employed at aerion. We just happen to use it at our company and I was missing a CLI for it.

I was frustrated with starting/stopping/resuming my timers through the [Aerion UI](https://www.aerion.app/). This repository provides an `aerion-cli` command to interact with aerion fom the command line.

With this tool you don't need to leave your development environment to update your aerion time entries. You can now do this:

```sh
$ aerion-cli start project1 "Task A"
# do some work
$ aerion-cli start project1 "Task B"
# do some work
$ aerion-cli start project2 "Task C"
# do some work
$ aerion-cli stop
# confirm that you've worked enough today
$ aerion-cli today
# close your laptop and drink a cup of tea on the couch
```

> [!NOTE]
> This application goes very well with my VSCode extension for aerion. Check it out here: https://marketplace.visualstudio.com/items?itemName=fischeversenker.aerion

## Installation

If you have `go` installed, you can install the `aerion-cli` like this:

```sh
go install github.com/fischeversenker/aerion-cli
```

If you don't have `go` installed, download a prebuilt binary from the latest [release](https://github.com/fischeversenker/aerion-cli/releases/latest) and put it somewhere in your `$PATH`.

## Usage

### Login

First you have to login:

```sh
aerion-cli login
```

This will ask for your company name, your username, and your password. The `aerion-cli` does not store the credentials that you provide here. It only stores the received token after a successful login. You only need to do this once.

### Today's time entries

Once you logged in, you can now get your time entries of today:

```sh
aerion-cli status # alias: today
```

This will produce something like this:

```sh
$ aerion-cli status
Project 1 | ⌛ 01h 15m | 📝 - My Comment for this time entry
Project 2 |    00h 45m | 📝 - Other Comment
total     |    02h 00m
```

> Add `--color` (or `-c`) to the command to get a more "colorful" output: `aerion-cli today -c`

### Yesterday's time entries

Similar to the `status`/`today` command, there is a `yesterday` command that shows the time entries of yesterday:

```sh
aerion-cli yesterday
```

### List projects

To get a list of all available projects:

```sh
aerion-cli projects list
```

This will print a list of active (non-complete) projects:

```sh
$ aerion-cli projects list
123      Project 1
124      Project 2
125      Project 3
```

### Set Project Aliases

To be able to start new time entries from the command line, you need to set up project aliases. Use the Project IDs that you got from the `aerion-cli projects list` command.

```sh
aerion-cli projects alias 123 proj1
```

Run this to see your currently configured project aliases:

```sh
aerion-cli projects alias
```

### Start a new time entry

Once you have aliases set up for your current project(s), you can start tracking your time like so:

```sh
aerion-cli start proj1
```

This will resume an existing time entry from today if there is one. It will stop any other running time entries. If there is no existing time entry it will start a new one.

You can add comments to your time entry like so:

```sh
aerion-cli start proj1 "Feature ABC"
```

Comments will be appended to any previous comment, so after running another `start` command like this:

```sh
aerion-cli start proj1 "Feature DEF"
```

you end up with a `today` like this:

```sh
$ aerion-cli today
proj1 | ⌛ 01h 22m | 📝 - Feature ABC - Feature DEF
```

## Help

Run this to get general help

```sh
aerion-cli help
```

or add a specific command to get more detailed help:

```sh
aerion-cli help start
```

## Usage tips

Add an alias for `aerion-cli` to your shell profile to make it easier to interact with aerion. I aliased `aerion-cli` to `cc` and `aerion-cli status --color` to `ccst` for very convenient workflows:

Add this to your `~/.bashrc` (or `~/.zshrc` or similar):
```
alias ac='aerion-cli'
alias acst='aerion-cli status -c'
```

