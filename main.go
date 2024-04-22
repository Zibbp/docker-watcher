package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zibbp/docker-watcher/config"
	"github.com/zibbp/docker-watcher/webhook"
)

var (
	CONFIG_PATH = "/data/config.json"
)

func main() {
	// setup logging
	debug := os.Getenv("DEBUG")
	if debug == "true" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	c, err := config.ReadConfig(CONFIG_PATH)
	if err != nil {
		log.Fatal().Err(err).Msg("error reading confg")
	}

	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating docker client")
	}
	defer apiClient.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// create a channel to signal goroutines to stop
	stopCh := make(chan bool)

	// monitor events for each container
	for _, container := range c.Containers {
		log.Info().Msgf("monitoring %s for events %s", container.Name, container.Events)
		go monitorContainerEvents(ctx, apiClient, container, stopCh)
	}

	// wait for an OS signal or a goroutine to stop
	select {
	case <-sigCh:
		log.Info().Msg("received stop signal, shutting down")
	case <-stopCh:
		log.Info().Msg("goroutine stopped, shutting down")
	}

	cancel()
}

func monitorContainerEvents(ctx context.Context, apiClient *client.Client, container config.Container, stopCh chan bool) {
	eventChan, errChan := apiClient.Events(ctx, types.EventsOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "container",
			Value: container.Name,
		}),
	})

	for {
		select {
		case event := <-eventChan:
			if contains(container.Events, event.Action) {

				// construct body to send
				var exitCodeText string
				exitCode := ""
				if val, ok := event.Actor.Attributes["exitCode"]; ok {
					exitCode = val
				}
				if event.Action == events.ActionDie {
					exitCodeText = fmt.Sprintf(" (exit code %s)", exitCode)
				}
				body := fmt.Sprintf("container %s: %s%s", container.Name, event.Action, exitCodeText)

				// log the event
				log.Debug().Msgf("[%s] Event: %#v", container.Name, event)
				log.Info().Msgf("%s", body)

				// send webhook
				err := webhook.SendWebhook(container.WebhookURL, body)
				if err != nil {
					log.Error().Err(err).Msg("error sending webhook")
				}
			}
		case err := <-errChan:
			log.Error().Err(err).Msg("received error from docker api")
			stopCh <- true
			return
		case <-ctx.Done():
			return
		}
	}
}

func contains(slice []events.Action, elem events.Action) bool {
	for _, v := range slice {
		if v == elem {
			return true
		}
	}
	return false
}
