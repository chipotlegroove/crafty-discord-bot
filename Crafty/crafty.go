// Package crafty handles Crafty Controller requests
package crafty

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	requests "github.com/chipotlegroove/crafty-discord-bot/Requests"
)

type ServerStatus int

var serverStatusName = map[ServerStatus]string{
	Offline:     "🔴 Offline",
	Online:      "🟢 Online",
	Unreachable: "💥 Unreachable",
}

func (ss ServerStatus) Label() string {
	return serverStatusName[ss]
}

const (
	Offline ServerStatus = iota
	Online
	Unreachable
)

type Server struct {
	ServerID   string `json:"server_id"`
	ServerName string `json:"server_name"`
	Status     ServerStatus
}

type ServersResponse struct {
	Data []Server `json:"data"`
}

type ServerStatusData struct {
	Running bool `json:"running"`
}

type ServerStatusResponse struct {
	Data ServerStatusData `json:"data"`
}

type ServerActionResponse struct {
	Status string `json:"status"`
}

var (
	token   string
	baseURL string
	client  *http.Client
	bearer  string
)

func Init(t string, url string) {
	token = t
	baseURL = url
	baseURL += "/api/v2"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client = &http.Client{Transport: tr}
	bearer = "Bearer " + token
}

func GetServersWithStatus() ([]Server, error) {
	if token == "" || baseURL == "" {
		return nil, errors.New("token or baseURL not set")
	}

	servers, err := getServers()
	if err != nil {
		return nil, fmt.Errorf("error getting servers: %w", err)
	}

	for i := range servers.Data {
		running, err := ServerIsRunning(servers.Data[i].ServerID)
		if err != nil {
			log.Printf("failed to get status for server %s: %s", servers.Data[i].ServerID, servers.Data[i].ServerName)
			servers.Data[i].Status = Unreachable
			continue
		}

		if running {
			servers.Data[i].Status = Online
		} else {
			servers.Data[i].Status = Offline
		}
	}

	return servers.Data, nil
}

func getServers() (res ServersResponse, err error) {
	requestURL := baseURL + "/servers"
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return ServersResponse{}, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", bearer)

	resp, err := client.Do(req)
	if err != nil {
		return ServersResponse{}, fmt.Errorf("error on response: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ServersResponse{}, fmt.Errorf("error while reading response: %w", err)
	}

	var formattedResponse ServersResponse
	json.Unmarshal(body, &formattedResponse)

	return formattedResponse, nil
}

func ServerIsRunning(serverID string) (bool, error) {
	requestURL := baseURL + "/servers/" + serverID + "/stats"

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", bearer)

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error on response: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("error while reading response: %w", err)
	}

	var formattedResponse ServerStatusResponse
	json.Unmarshal(body, &formattedResponse)

	return formattedResponse.Data.Running, nil
}

func ExecuteAction(serverName string, action string) error {
	servers, err := getServers()
	if err != nil {
		return fmt.Errorf("error getting servers: %w", err)
	}

	serverToStart := Server{}

	for _, server := range servers.Data {
		if server.ServerName == serverName {
			serverToStart = server
			break
		}
	}

	if serverToStart == (Server{}) {
		return fmt.Errorf("server not found")
	}

	requestURL := baseURL + "/servers/" + serverToStart.ServerID + "/action/" + action

	responseBody, err := requests.ExecuteRequest("POST", requestURL, bearer)
	if err != nil {
		return fmt.Errorf("error while executing requests: %w", err)
	}

	var formattedResponse ServerActionResponse
	json.Unmarshal(responseBody, &formattedResponse)

	if formattedResponse.Status != "ok" {
		return fmt.Errorf("error while executing action: %s", formattedResponse.Status)
	}

	return nil
}
