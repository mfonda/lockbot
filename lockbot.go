package main

import (
	"encoding/json"
	"github.com/mfonda/slash"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type lock struct {
	Key         string
	Username    string
	CreatedDate time.Time
	Notes       string
}

var lockfile string = ".lockbot.json"
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

	if len(os.Args) > 1 {
		lockfile = os.Args[1]
		lockJson, err := ioutil.ReadFile(lockfile)
		if err != nil {
			log.Fatalf("Failed to read '%s': %s", lockfile, err)
		}
		err = json.Unmarshal(lockJson, &locks)
		if err != nil {
			log.Fatalf("Failed to parse '%s': %s", lockfile, err)
		}
	}

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
		Key:         key,
		Username:    req.UserName,
		CreatedDate: time.Now().In(tz),
		Notes:       notes,
	}
	persist()

	return reply(locks[key].String())
}

func unlockHandler(req *slash.Request) (*slash.Response, error) {
	key, _ := parseCommand(req.Text)
	currentLock, exists := locks[key]
	if !exists {
		return reply(key + " is already unlocked")
	}
	delete(locks, key)
	persist()

	resp := "Successfully unlocked " + key
	if req.UserName != currentLock.Username {
		resp += " (cc @" + currentLock.Username + ")"
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
	date := l.CreatedDate.Format("Mon Jan 2 15:04:05 MST")
	desc := l.Key + " locked by @" + l.Username + " on " + date
	if len(l.Notes) > 0 {
		desc += " (" + l.Notes + ")"
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

func persist() error {
	lockJson, err := json.Marshal(locks)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(lockfile, lockJson, 0644)
}
