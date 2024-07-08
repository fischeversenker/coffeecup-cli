package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// returns accesstoken, refreshtoken, and error
func LoginWithPassword(company string, username string, password string) (string, string, error) {
	reqBody := url.Values{
		"grant_type": []string{"password"},
		"username":   []string{username},
		"password":   []string{password},
	}

	req, err := http.NewRequest("POST", "https://api.coffeecupapp.com/oauth2/token", strings.NewReader(reqBody.Encode()))
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Authorization", "Basic ZW1iZXI6cHVibGlj")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("companyurl", strings.Join([]string{"https://", company, ".coffeecup.app"}, ""))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		panic(err)
	}

	return responseBody["access_token"].(string), responseBody["refresh_token"].(string), nil
}

func GetUserId() (int, error) {
	req, err := http.NewRequest("GET", "https://api.coffeecupapp.com/v1/users/me", nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+GetAccessTokenFromConfig())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	type UserResponse struct {
		User struct {
			Id int
		}
	}

	var responseBody UserResponse
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		panic(err)
	}

	return responseBody.User.Id, nil
}

func LoginWithRefreshToken() (string, string, error) {
	reqBody := url.Values{
		"grant_type":    []string{"refresh_token"},
		"refresh_token": []string{GetRefreshTokenFromConfig()},
	}

	req, err := http.NewRequest("POST", "https://api.coffeecupapp.com/oauth2/token", strings.NewReader(reqBody.Encode()))
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Authorization", "Basic ZW1iZXI6cHVibGlj")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		panic(err)
	}

	return responseBody["access_token"].(string), responseBody["refresh_token"].(string), nil
}

type Project struct {
	Id   int
	Name string
}

type ProjectsResponse struct {
	Projects []Project
	Meta     struct {
		Total int
	}
	Status int
}

func GetProjects() ([]Project, error) {
	req, err := http.NewRequest("GET", "https://api.coffeecupapp.com/v1/projects", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+GetAccessTokenFromConfig())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var projectsResponse ProjectsResponse
	err = json.NewDecoder(resp.Body).Decode(&projectsResponse)
	if err != nil {
		return nil, err
	}
	if projectsResponse.Status == 401 {
		return nil, fmt.Errorf("unauthorized")
	}

	projects := make([]Project, int(projectsResponse.Meta.Total))
	for i, p := range projectsResponse.Projects {
		projects[i] = Project{p.Id, p.Name}
	}
	return projects, nil
}

type TimeEntry struct {
	Id        int
	Project   int
	Task      int
	Team      int
	User      int
	Comment   string
	Running   bool
	CreatedAt string
	Day       string
	Duration  int
}

type TimeEntriesResponse struct {
	TimeEntries []TimeEntry
	Meta        struct {
		Total int
	}
}

func GetTodaysTimeEntries() ([]TimeEntry, error) {
	userId := strconv.Itoa(GetUserIdFromConfig())
	today := time.Now().Format("2006-01-02")
	url := "https://api.coffeecupapp.com/v1/timeentries?limit=1000&where={\"user\":\"" + userId + "\",\"day\":\"" + today + "\"}&sort=day%20ASC,sorting%20ASC"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+GetAccessTokenFromConfig())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var timeEntries TimeEntriesResponse
	err = json.NewDecoder(resp.Body).Decode(&timeEntries)
	if err != nil {
		return nil, err
	}

	// sort time entries (timeEntries.TimeEntries) by CreatedAt
	sort.Slice(timeEntries.TimeEntries, func(i, j int) bool {
		return timeEntries.TimeEntries[i].CreatedAt > timeEntries.TimeEntries[j].CreatedAt
	})

	return timeEntries.TimeEntries, nil
}
