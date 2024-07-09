package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
	Id           int    `json:"id"`
	ProjectId    int    `json:"project"`
	TaskId       int    `json:"task"`
	TeamId       int    `json:"team"`
	UserId       int    `json:"user"`
	Comment      string `json:"comment"`
	Running      bool   `json:"running"`
	CreatedAt    string `json:"createdAt"`
	Day          string `json:"day"`
	Duration     int    `json:"duration"`
	Sorting      int    `json:"sorting"`
	TrackingType string `json:"trackingType"`
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

	type TimeEntriesResponse struct {
		TimeEntries []TimeEntry `json:"timeEntries"`
		Meta        struct {
			Total int `json:"total"`
		} `json:"Meta"`
		Status int    `json:"status"`
		Error  string `json:"error"`
		Raw    string `json:"raw"`
	}

	var timeEntriesResponse TimeEntriesResponse
	err = json.NewDecoder(resp.Body).Decode(&timeEntriesResponse)
	if err != nil {
		return nil, err
	}
	if timeEntriesResponse.Status == 401 {
		return nil, fmt.Errorf("unauthorized")
	}
	if timeEntriesResponse.Error != "" {
		return nil, fmt.Errorf(timeEntriesResponse.Raw)
	}

	return timeEntriesResponse.TimeEntries, nil
}

func batchUpdateTimeEntries(timeEntriesToBeUpdated []TimeEntry) error {
	url := "https://api.coffeecupapp.com/v1/timeEntries/batchUpdate"
	type TimeEntryUpdate struct {
		TimeEntries []TimeEntry `json:"timeEntries"`
	}

	timeEntryToBeUpdated := TimeEntryUpdate{
		TimeEntries: make([]TimeEntry, 0),
	}
	timeEntryToBeUpdated.TimeEntries = append(timeEntryToBeUpdated.TimeEntries, timeEntriesToBeUpdated...)
	payload, err := json.Marshal(timeEntryToBeUpdated)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", string(payload))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+GetAccessTokenFromConfig())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var timeEntriesBatchUpdateResponse struct {
		Status int `json:"status"`
	}
	err = json.NewDecoder(resp.Body).Decode(&timeEntriesBatchUpdateResponse)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", timeEntriesBatchUpdateResponse)
	if timeEntriesBatchUpdateResponse.Status == 401 {
		return fmt.Errorf("unauthorized")
	}

	return nil
}

func UpdateTimeEntry(timeEntry TimeEntry) error {
	url := "https://api.coffeecupapp.com/v1/timeEntries/" + strconv.Itoa(timeEntry.Id)

	type TimeEntryUpdate struct {
		TimeEntry TimeEntry `json:"timeEntry"`
	}

	timeEntryToBeUpdated := TimeEntryUpdate{
		TimeEntry: timeEntry,
	}
	payload, err := json.Marshal(timeEntryToBeUpdated)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+GetAccessTokenFromConfig())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var timeEntryUpdateResponse struct {
		Status int    `json:"status"`
		Raw    string `json:"raw"`
		Error  string `json:"error"`
	}
	err = json.NewDecoder(resp.Body).Decode(&timeEntryUpdateResponse)
	if err != nil {
		return err
	}
	if timeEntryUpdateResponse.Status == 401 {
		return fmt.Errorf("unauthorized")
	}
	if timeEntryUpdateResponse.Error != "" {
		return fmt.Errorf(timeEntryUpdateResponse.Raw)
	}

	return nil
}

type NewTimeEntry struct {
	ProjectId    int    `json:"project"`
	Comment      string `json:"comment"`
	Day          string `json:"day"`
	Running      bool   `json:"running"`
	Duration     int    `json:"duration"`
	Sorting      int    `json:"sorting"`
	TaskId       int    `json:"task"`
	TrackingType string `json:"trackingType"`
	TeamId       int    `json:"team"`
	UserId       int    `json:"user"`
}

func CreateTimeEntry(timeEntry NewTimeEntry) error {
	url := "https://api.coffeecupapp.com/v1/timeEntries"

	type TimeEntryCreation struct {
		TimeEntry NewTimeEntry `json:"timeEntry"`
	}

	timeEntryToBeCreated := TimeEntryCreation{
		TimeEntry: timeEntry,
	}
	payload, err := json.Marshal(timeEntryToBeCreated)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+GetAccessTokenFromConfig())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var timeEntriesCreationResponse struct {
		Status int    `json:"status"`
		Raw    string `json:"raw"`
		Error  string `json:"error"`
	}
	err = json.NewDecoder(resp.Body).Decode(&timeEntriesCreationResponse)
	if err != nil {
		return err
	}
	if timeEntriesCreationResponse.Status == 401 {
		return fmt.Errorf("unauthorized")
	}
	if timeEntriesCreationResponse.Error != "" {
		return fmt.Errorf(timeEntriesCreationResponse.Raw)
	}

	return nil
}
