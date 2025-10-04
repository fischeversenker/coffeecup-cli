package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type WorklogEntry struct {
	Id       int
	Duration int
}

const (
	WorklowFolderPath = ".local/state/aerion"
	WorklogFileName   = "worklog.log"
)

func GetWorklogPath() string {
	return filepath.Join(os.Getenv("HOME"), WorklowFolderPath, WorklogFileName)
}

func ReadWorklog() ([]WorklogEntry, error) {
	err := os.MkdirAll(ConfigFolderPath, os.ModePerm)
	if err != nil {
		return []WorklogEntry{}, err
	}

	readFile, err := os.Open(GetWorklogPath())
	if err != nil {
		return []WorklogEntry{}, err
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	var worklogList []WorklogEntry
	for fileScanner.Scan() {
		line := fileScanner.Text()
		lineParts := strings.Split(line, "=")
		if len(lineParts) != 2 {
			fmt.Printf("Couldn't split line %s from worklog.log\n", line)
		} else {
			worklogId, _ := strconv.Atoi(lineParts[0])
			worklogDuration, _ := strconv.Atoi(lineParts[1])
			worklogList = append(
				worklogList,
				WorklogEntry{Id: worklogId, Duration: worklogDuration},
			)
		}
	}

	return worklogList, nil
}

func GetWorklogEntry(id int) (WorklogEntry, error) {
	worklog, err := ReadWorklog()
	if err != nil {
		return WorklogEntry{-1, 0}, err
	}
	for _, worklogEntry := range worklog {
		if id == worklogEntry.Id {
			return worklogEntry, nil
		}
	}
	return WorklogEntry{-1, 0}, nil
}

func UpsertWorklogEntry(worklogEntry WorklogEntry) error {
	err := os.MkdirAll(filepath.Join(os.Getenv("HOME"), WorklowFolderPath), os.ModePerm)
	if err != nil {
		return err
	}

	existingWorklog, err := ReadWorklog()
	newWorklog := []string{}
	wasExisting := false
	for _, otherWorklogEntry := range existingWorklog {
		if worklogEntry.Id == otherWorklogEntry.Id {
			newWorklog = append(newWorklog, worklogEntryToString(worklogEntry))
			wasExisting = true
		} else {
			newWorklog = append(newWorklog, worklogEntryToString(otherWorklogEntry))
		}
	}
	if !wasExisting {
		newWorklog = append(newWorklog, worklogEntryToString(worklogEntry))
	}

	err = os.WriteFile(GetWorklogPath(), []byte(strings.Join(newWorklog, "")), 0644)
	if err != nil {
		return err
	}

	return nil
}

func worklogEntryToString(worklogEntry WorklogEntry) string {
	return strconv.Itoa(worklogEntry.Id) + "=" + strconv.Itoa(worklogEntry.Duration) + "\n"
}
