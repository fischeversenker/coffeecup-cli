# Coffeecup CLI

> [!NOTE]
> This is a private project! I'm not employed at Coffeecup. We just happen to use it at our company and I was missing a CLI for it.

I was frustrated with starting/stopping/resuming my timers through the [Coffeecup UI](https://www.coffeecup.app/). This repository provides a `coffeecup` CLI command to interact with Coffeecup fom the command line.

With this tool you don't need to leave your development environment to update your Coffeecup time entries. You can now do this:

```sh
$ coffeecup start project1 "Task A"
# do some work
$ coffeecup start project1 "Task B"
# do some work
$ coffeecup start project2 "Task C"
# do some work
$ coffeecup stop
# drink beer
```

## Installation

If you have `go` installed, you can install the `coffeecup` cli like this:

```sh
go install github.com/fischeversenker/coffeecup
```

If you don't have `go` installed, download a prebuilt binary from the latest [release](https://github.com/fischeversenker/coffeecup/releases/latest) and put it somewhere in your `$PATH`.

## Usage

### Login

First you have to login:

```sh
coffeecup login
```

This will ask for your company name, your username, and your password. The `coffeecup` CLI does not store the credentials that you provide here. It only stores the received token after a successful login. You only need to do this once.

### Today's time entries

Once you logged in, you can now get your time entries of today:

```sh
coffeecup today
```

This will produce something like this:

```sh
$ coffeecup today
Project 1 | ⌛ 01h 15m | 📝 - My Comment for this time entry
Project 2 |    00h 45m | 📝 - Other Comment
total     |    02h 00m
```

> Add `-c` to the command to get a more "colorful" output: `coffeecup today -c`

### List projects

To get a list of all available projects:

```sh
coffeecup projects list
```

This will print a list of active (non-complete) projects:

```sh
$ coffeecup projects list
123      Project 1
124      Project 2
125      Project 3
```

### Set Project Aliases

To be able to start new time entries from the command line, you need to set up project aliases. Use the Project IDs that you got from the `coffeecup projects list` command.

```sh
coffeecup projects alias 123 proj1
```

Run this to see your currently configured project aliases:

```sh
coffeecup projects alias
```

### Start a new time entry

Once you have aliases set up for your current project(s), you can start tracking your time like so:

```sh
coffeecup start proj1
```

This will resume an existing time entry from today if there is one. It will stop any other running time entries. If there is no existing time entry it will start a new one.

You can add comments to your time entry like so:

```sh
coffeecup start proj1 "Feature ABC"
```

Comments will be appended to any previous comment, so after running another `start` command like this:

```sh
coffeecup start proj1 "Feature DEF"
```

you end up with a `today` like this:

```sh
$ coffeecup today
proj1 | ⌛ 01h 22m | 📝 - Feature ABC - Feature DEF
```

## Help

Run this to get general help

```sh
coffeecup help
```

or add a specific command to get more detailed help:

```sh
coffeecup help start
```

# Development

Coffeecup API Docs:
https://dev.coffeecupapp.com/
