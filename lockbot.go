package main

import (
	"github.com/mfonda/slash"
	"log"
	"os"
	"strings"
	"time"
)

type lock struct {
	key         string
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
		resp := "Error: already locked\n"
		resp += currentLock.String()
		return slash.NewInChannelResponse(resp, nil), nil
	}

	tz, _ := time.LoadLocation("America/Los_Angeles")
	locks[key] = lock{
		key:         key,
		username:    req.UserName,
		createdDate: time.Now().In(tz),
		notes:       notes,
	}

	return slash.NewInChannelResponse(locks[key].String(), nil), nil
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
		for _, l := range locks {
			resp += l.String() + "\n"
		}
		return slash.NewInChannelResponse(resp, nil), nil
	}
	if !exists {
		return slash.NewInChannelResponse(key+" is currently unlocked", nil), nil
	}

	return slash.NewInChannelResponse(currentLock.String(), nil), nil
}

func (l lock) String() string {
	date := l.createdDate.Format("Mon Jan 2 15:04:05 MST")
	desc := l.key + " locked by @" + l.username + " on " + date
	if len(l.notes) > 0 {
		desc += " (" + l.notes + ")"
	}
	return desc
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
