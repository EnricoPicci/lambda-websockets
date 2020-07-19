package main

import (
	"context"
	"log"
	"net/http"

	"github.com/websockets-lambda/server/mongostore"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type connectionStorer interface {
	GetConnectionIDs(ctx context.Context) ([]string, error)
	AddConnectionID(ctx context.Context, connectionID string) error
	MarkConnectionIDDisconnected(ctx context.Context, connectionID string) error
}

var connectionStore connectionStorer

func handleRequest(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Lambda Handle Request started")

	if connectionStore == nil {
		connectionStore = mongostore.NewMongoStore(ctx)
	}

	rc := event.RequestContext
	switch rk := rc.RouteKey; rk {
	case "$connect":
		log.Println("Connect", rc.ConnectionID)
		err := connectionStore.AddConnectionID(ctx, rc.ConnectionID)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, err
		}
		break
	case "$disconnect":
		log.Println("Disconnect", rc.ConnectionID)
		err := connectionStore.MarkConnectionIDDisconnected(ctx, rc.ConnectionID)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, err
		}
		break
	case "$default":
		log.Println("Default", rc.ConnectionID, event.Body)
		err := echo(ctx, event, connectionStore)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, err
		}
		break
	default:
		log.Fatalf("Unknown RouteKey %v", rk)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	log.Println("main starts")
	lambda.Start(handleRequest)
	log.Println("main ends - this line seems not to be written in the log")
}
