# Slack Event Mediator

Slack Event Mediator is a simple Cloud Run function that listens to [Slack events](https://api.slack.com/events?filter=Events) and forwards them to [Cloud Pub/Sub](https://cloud.google.com/pubsub).

## Environment Variables

* Required
  * `GOOGLE_CLOUD_PROJECT`
  * `SLACK_SIGNING_SECRET`
* Optional
  * `PUBSUB_TOPIC` (defualt: `slack-events`)

## Deployment

Clone this repository then deploy it to Cloud Run functions. Example:

```bash
FUNCTION_NAME=slack-event-mediator
REGION=asia-northeast1
GOOGLE_CLOUD_PROJECT=YOUR_PROJECT
SECRET=YOUR_SECRET_MANAGER_NAME

gcloud functions deploy "$FUNCTION_NAME" \
  --gen2 \
  --runtime=go123 \
  --region="$REGION" \
  --source=. \
  --entry-point=SlackEventMediator \
  --trigger-http \
  --allow-unauthenticated \
  --set-env-vars=GOOGLE_CLOUD_PROJECT="$GOOGLE_CLOUD_PROJECT" \
  --set-secrets=SLACK_SIGNING_SECRET="${SECRET}:latest"
```
