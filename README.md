# Docker Watcher

Get notified on specific docker container events using webhooks.

## Config

The container expects a config mounted to `/data/config.json` in the container.

The config contains an array of containers to watch with the following options.

- `name`: Name of the container (or id).
- `webhook_url`: URL of a webhook to send the notification to (I've only tested [ntfy.sh](ntfy.sh)).
- `events`: Array of events to monitor. See https://pkg.go.dev/github.com/docker/docker/api/types/events#Action for a full list

See [config.example.json](config.example.json) for an example

## Getting Started

1. Copy the `compose.yml` file
2. Create the local `data` directory and create a `config.json` inside
   1. See [config.example.json](config.example.json) for an example
