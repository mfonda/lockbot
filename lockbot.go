package main

import (
	"github.com/mfonda/slash"
	"log"
	"os"
	"strings"
	"time"
)

type lock struct {
	username    string
	createdDate time.Time
	notes       string
}

var locks map[string]lock

func main() {
	required := [4]string{"SLACK_VERIFICATION_TOKEN", "SLACK_SSL_CERT_PATH", "SLACK_SSL_KEY_PATH", "SLACK_PORT"}
	for _, envVar := range required {
		_, exists := os.LookupEnv(envVar)
		if !exists {
			log.Fatal("Missing required environment var " + envVar)
		}
	}
	token := os.Getenv("SLACK_VERIFICATION_TOKEN")
	certFile := os.Getenv("SLACK_SSL_CERT_PATH")
	keyFile := os.Getenv("SLACK_SSL_KEY_PATH")
	port := os.Getenv("SLACK_PORT")

	slash.HandleFunc("/lock", token, lockHandler)
	slash.HandleFunc("/unlock", token, unlockHandler)
	slash.HandleFunc("/lock-status", token, statusHandler)
	locks = make(map[string]lock)
	log.Fatal(slash.ListenAndServeTLS(":"+port, certFile, keyFile))
}

func lockHandler(req *slash.Request) (*slash.Response, error) {
	key, notes := parseCommand(req.Text)
	currentLock, exists := locks[key]
	if exists {
		resp := "Error: '" + key + "' is currently locked (@" + currentLock.username + " at " + currentLock.createdDate.String() + ")"
		return slash.NewInChannelResponse(resp, nil), nil
	}

	locks[key] = lock{
		username:    req.UserName,
		createdDate: time.Now(),
		notes:       notes,
	}

	return slash.NewInChannelResponse("Successfully locked "+key, nil), nil
}

func unlockHandler(req *slash.Request) (*slash.Response, error) {
	key, _ := parseCommand(req.Text)
	currentLock, exists := locks[key]
	if !exists {
		return slash.NewInChannelResponse(key+" is already unlocked", nil), nil
	}
	delete(locks, key)

	resp := "Successfully unlocked " + key
	if req.UserName != currentLock.username {
		resp += " (cc @" + currentLock.username + ")"
	}
	return slash.NewInChannelResponse(resp, nil), nil
}

func statusHandler(req *slash.Request) (*slash.Response, error) {
	key, _ := parseCommand(req.Text)
	currentLock, exists := locks[key]
	if key == "" {
		if len(locks) == 0 {
			return slash.NewInChannelResponse("Everything is unlocked!", nil), nil
		}
		resp := "Current locks:\n"
		for lockKey, lockInfo := range locks {
			resp += lockKey + " by @" + lockInfo.username + " (" + lockInfo.createdDate.String() + ")\n"
		}
		return slash.NewInChannelResponse(resp, nil), nil
	}
	if !exists {
		return slash.NewInChannelResponse(key+" is currently unlocked", nil), nil
	}
	resp := key + " was locked by " + currentLock.username + " at " + currentLock.createdDate.String()
	if currentLock.notes != "" {
		resp += ": " + currentLock.notes
	}
	return slash.NewInChannelResponse(resp, nil), nil
}

func parseCommand(command string) (string, string) {
	parts := strings.SplitN(command, " ", 2)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}