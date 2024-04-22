package webhook

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func SendWebhook(url, body string) error {
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(body))
	if err != nil {
		return err
	}

	client := &http.Client{}

	// retry up to 3 times
	for i := 0; i < 3; i++ {
		if i > 0 {
			log.Info().Msgf("retrying request, attempt %d", i+1)
			time.Sleep(2 * time.Second)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("error sending request")
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		return nil
	}

	return fmt.Errorf("maximum number of retries exceeded")
}
