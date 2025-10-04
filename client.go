package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
	Raw          string `json:"raw"`
	Status       int    `json:"status"`
}

var publicApiToken = "Basic " + base64.StdEncoding.EncodeToString([]byte("coffeecup-cli:public"))

func GetApiBaseUrl() (string, error) {
	cfg, err := ReadConfig()
	if err != nil {
		return "", err
	}

	if cfg.User.Company == "" {
		return "", fmt.Errorf("No company set. Are you logged in? Please run the login command first.")
	}
	return "https://" + cfg.User.Company + ".aerion.app", nil
}

func EnsureLoggedIn() error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}

	if cfg.User.ExpiresAt > time.Now().Unix() {
		return nil
	}

	return LoginWithRefreshToken()
}

// returns accesstoken, refreshtoken, and error
func LoginWithPassword(username string, password string) error {
	reqBody := url.Values{
		"grant_type": []string{"password"},
		"username":   []string{username},
		"password":   []string{password},
	}

	apiBaseURL, err := GetApiBaseUrl()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiBaseURL+"/oauth2/token", strings.NewReader(reqBody.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", publicApiToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("companyurl", apiBaseURL)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var responseBody TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		panic(err)
	}

	if responseBody.Error != "" {
		return fmt.Errorf(responseBody.Raw)
	}

	StoreTokens(responseBody.AccessToken, responseBody.RefreshToken, responseBody.ExpiresIn)

	return nil
}

type User struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

type UserResponse struct {
	User User
}

func GetUser() (User, error) {
	apiBaseURL, err := GetApiBaseUrl()
	if err != nil {
		return User{}, err
	}

	req, err := http.NewRequest("GET", apiBaseURL+"/v1/users/me", nil)
	if err != nil {
		return User{}, err
	}

	req.Header.Set("Authorization", "Bearer "+GetAccessTokenFromConfig())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return User{}, err
	}
	if resp.StatusCode != 200 {
		return User{}, fmt.Errorf("unauthorized")
	}
	defer resp.Body.Close()

	var responseBody UserResponse
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		panic(err)
	}

	return responseBody.User, nil
}

func LoginWithRefreshToken() error {
	refreshToken := GetRefreshTokenFromConfig()
	if refreshToken == "" {
		return fmt.Errorf("no refresh token found")
	}
	reqBody := url.Values{
		"grant_type":    []string{"refresh_token"},
		"refresh_token": []string{refreshToken},
	}

	apiBaseURL, err := GetApiBaseUrl()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiBaseURL+"/oauth2/token", strings.NewReader(reqBody.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", publicApiToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var responseBody TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		panic(err)
	}
	if responseBody.Error != "" {
		return fmt.Errorf(responseBody.Raw)
	}

	StoreTokens(responseBody.AccessToken, responseBody.RefreshToken, responseBody.ExpiresIn)
	return nil
}

type Project struct {
	Id   int
	Name string
}

type ProjectsResponse struct {
	Projects []Project `json:"projects"`
	Meta     struct {
		Total int `json:"total"`
	} `json:"Meta"`
	Status int    `json:"status"`
	Error  string `json:"error"`
	Raw    string `json:"raw"`
}

func GetProjects() ([]Project, error) {
	apiBaseURL, err := GetApiBaseUrl()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", apiBaseURL+"/v1/projects?status=1", nil)
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
	if projectsResponse.Error != "" {
		return nil, fmt.Errorf(projectsResponse.Raw)
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

type TimeEntriesResponse struct {
	TimeEntries []TimeEntry `json:"timeEntries"`
	Meta        struct {
		Total int `json:"total"`
	} `json:"Meta"`
	Status int    `json:"status"`
	Error  string `json:"error"`
	Raw    string `json:"raw"`
}

func GetTodaysTimeEntries() ([]TimeEntry, error) {
	today := time.Now().Format("2006-01-02")
	return getTimeEntriesForDay(today)
}

// this actually returns the time entries of the previous working day.
// it doesn't need to be yesterday, could be last Friday if it's a Monday today.
func GetYesterdaysTimeEntries() ([]TimeEntry, error) {
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)
	yesterdaysWeekday := yesterday.Weekday()
	for ; yesterdaysWeekday == time.Saturday || yesterdaysWeekday == time.Sunday; yesterdaysWeekday = yesterday.Weekday() {
		yesterday = yesterday.AddDate(0, 0, -1)
	}
	yesterdayFormatted := yesterday.Format("2006-01-02")
	return getTimeEntriesForDay(yesterdayFormatted)
}

func getTimeEntriesForDay(day string) ([]TimeEntry, error) {
	userId := strconv.Itoa(GetUserIdFromConfig())
	apiBaseURL, err := GetApiBaseUrl()
	if err != nil {
		return nil, err
	}

	url := apiBaseURL + "/v1/timeentries?limit=1000&user=" + userId + "&day=" + day + "&sort=day%20ASC,sorting%20ASC"
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

func GetLastTimeEntryForProject(projectId int) (TimeEntry, error) {
	userId := strconv.Itoa(GetUserIdFromConfig())
	apiBaseURL, err := GetApiBaseUrl()
	if err != nil {
		return TimeEntry{}, err
	}

	url := apiBaseURL + "/v1/timeentries?limit=1&user=" + userId + "&project=" + strconv.Itoa(projectId) + "&sort=day%20DESC"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return TimeEntry{}, err
	}

	req.Header.Set("Authorization", "Bearer "+GetAccessTokenFromConfig())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return TimeEntry{}, err
	}
	defer resp.Body.Close()

	var timeEntriesResponse TimeEntriesResponse
	err = json.NewDecoder(resp.Body).Decode(&timeEntriesResponse)
	if err != nil {
		return TimeEntry{}, err
	}

	if timeEntriesResponse.Status == 401 {
		return TimeEntry{}, fmt.Errorf("unauthorized")
	}
	if timeEntriesResponse.Error != "" {
		return TimeEntry{}, fmt.Errorf(timeEntriesResponse.Raw)
	}
	if timeEntriesResponse.Meta.Total == 0 {
		return TimeEntry{}, fmt.Errorf("no time entries found")
	}

	return timeEntriesResponse.TimeEntries[0], nil
}

func UpdateTimeEntry(timeEntry TimeEntry) error {
	apiBaseURL, err := GetApiBaseUrl()
	if err != nil {
		return err
	}

	url := apiBaseURL + "/v1/timeEntries/" + strconv.Itoa(timeEntry.Id)

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
	UserId       int    `json:"user"`
}

func CreateTimeEntry(timeEntry NewTimeEntry) error {
	apiBaseURL, err := GetApiBaseUrl()
	if err != nil {
		return err
	}

	url := apiBaseURL + "/v1/timeEntries"

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
