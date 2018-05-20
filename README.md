# lockbot

`lockbot` is a Slack bot for informal management of shared resources amongst a team. `lockbot` does not affect
the outside world in any way (i.e. it does not actually "lock" any shared resources), it simply keeps track of
what users are currently using what resources. It's meant to bridge the gap between situations where just asking
around is too messy but building formal systems is overkill.

This can be useful for day-to-day tasks that might not require formal structure, such as who is currently using
a shared server or who currently has the only physical copy of a thing in the office, say a book or a tool, etc.

`lockbot` provides no structure beyond simply marking resources as in-use; use it as you best see fit.

## Usage

`lockbot` provides handlers for the following Slash commands:

```
/lock <resource name> [notes]
```
Locks the given resource, optionally including notes about why it was locked

```
/unlock <resource name>
```
Unlocks the given resource

```
/locks
```
Lists all resources currently locked

## Installation

```
go get github.com/mfonda/lockbot
```

## Configuration

`lockbot` requires the following environment variables to be set:

 - `SLACK_VERIFICATION_TOKEN`: token provided via Slack to verify that requests are indeed coming from Slack
 - `SLACK_SSL_CERT_PATH`: path to SSL certificate file (Slack requires slack commands to run via SSL)
 - `SLACK_SSL_KEY_PATH`: path to SSL key file
 - `SLACK_PORT`: port the service should listen on (e.g. `443`)

On the Slack side of things, add Slash commands for the `/lock`, `/unlock`, and `/locks`:

Example for `/lock`

 - `Command`: `/lock`
 - `Request URL`: `https://example.com/lock` (include port in URL if it's anything other than `443`)

By default, `lockbot` will persist locks to disk in `.lockbot.json`. A custom path may be specified by
passing it in as the first argument when running `lockbot`:

```
./lockbot [path to lock json]
```
