package slackeventmediator

import (
	"cmp"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

var (
	topic         *pubsub.Topic
	signingSecret = os.Getenv("SLACK_SIGNING_SECRET")
)

const (
	defualtTopic = "slack-events"
)

func init() {
	functions.HTTP("SlackEventMediator", slackEventMediator)

	// Create a new Pub/Sub client.
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	topic = client.Topic(cmp.Or(os.Getenv("PUBSUB_TOPIC"), defualtTopic))
}

// slackEventMediator is an HTTP Cloud Function that processes incoming requests.
func slackEventMediator(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if _, err := sv.Write(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := sv.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
		w.WriteHeader(http.StatusOK)
		return
	}

	// All apps must acknowledge the event within 3 seconds.
	// https://api.slack.com/interactivity/handling#acknowledgment_response
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	res := topic.Publish(ctx, &pubsub.Message{
		Data: body,
		Attributes: map[string]string{
			"event_type":    eventsAPIEvent.Type,
			"team_id":       eventsAPIEvent.TeamID,
			"api_app_id":    eventsAPIEvent.APIAppID,
			"enterprise_id": eventsAPIEvent.EnterpriseID,
		},
	})
	if _, err := res.Get(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
