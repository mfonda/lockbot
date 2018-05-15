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

func init() {
	locks = make(map[string]lock)
}

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
	slash.HandleFunc("/locks", token, statusHandler)
	log.Fatal(slash.ListenAndServeTLS(":"+port, certFile, keyFile))
}

func lockHandler(req *slash.Request) (*slash.Response, error) {
	key, notes := parseCommand(req.Text)
	currentLock, exists := locks[key]
	if exists {
		resp := "Error: already locked\n"
		resp += currentLock.String()
		return reply(resp)
	}

	tz, _ := time.LoadLocation("America/Los_Angeles")
	locks[key] = lock{
		key:         key,
		username:    req.UserName,
		createdDate: time.Now().In(tz),
		notes:       notes,
	}

	return reply(locks[key].String())
}

func unlockHandler(req *slash.Request) (*slash.Response, error) {
	key, _ := parseCommand(req.Text)
	currentLock, exists := locks[key]
	if !exists {
		return reply(key + " is already unlocked")
	}
	delete(locks, key)

	resp := "Successfully unlocked " + key
	if req.UserName != currentLock.username {
		resp += " (cc @" + currentLock.username + ")"
	}
	return reply(resp)
}

func statusHandler(req *slash.Request) (*slash.Response, error) {
	if len(locks) == 0 {
		return reply("Everything is unlocked!")
	}

	resp := "Current locks:\n"
	for _, l := range locks {
		resp += l.String() + "\n"
	}
	return reply(resp)
}

func (l lock) String() string {
	date := l.createdDate.Format("Mon Jan 2 15:04:05 MST")
	desc := l.key + " locked by @" + l.username + " on " + date
	if len(l.notes) > 0 {
		desc += " (" + l.notes + ")"
	}
	return desc
}

func reply(text string) (*slash.Response, error) {
	return slash.NewInChannelResponse(text, nil), nil
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
