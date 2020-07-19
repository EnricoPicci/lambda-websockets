package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

func echoGoroutine(ctx context.Context, event events.APIGatewayWebsocketProxyRequest, store connectionStorer) error {
	sess, err := session.NewSession()
	if err != nil {
		log.Fatalln("Unable to create AWS session", err.Error())
	}
	dname := event.RequestContext.DomainName
	stage := event.RequestContext.Stage
	endpoint := fmt.Sprintf("https://%v/%v", dname, stage)
	apigateway := apigatewaymanagementapi.New(sess, aws.NewConfig().WithEndpoint(endpoint))

	body := event.Body
	resp := fmt.Sprintf("Echo me: %v", body)

	// if the body contains an integer, than a delay in the response is introduced
	delay, err := strconv.Atoi(body)
	if err != nil {
		delay = 0
	}

	// we simulate a long process via time.Sleep and we use a goroutine with a channel to signal the end of the processing
	// even if we use a goroutine, this lambda "instance" can process only one request at the time, i.e. if another client
	// sends a message while this function is in Sleep mode, another execution instance of the same function is triggered
	// with another execution context - for this reason we can not cache connectionIDs in memory, since we can have more
	// than one exectuion instance of the same Lambda function running at the same time - if we use a cache, the exectuion
	// context which started first would not know the existence of the connectionID which started the second execution
	// and therefore would have a connectionID cache not complete - for this reason we need always to read from the DB
	// the list of connectionIDs if we want to be sure that we have the complete list of them
	c := make(chan bool)
	go func() {
		time.Sleep(time.Duration(delay) * time.Second)
		c <- true
	}()

	select {
	case <-ctx.Done():
		err := ctx.Err()
		log.Println("request cancelled:", err)
		return err
	case <-c:
		connections, err := store.GetConnectionIDs(ctx)
		if err != nil {
			log.Fatalln("Unable to get connections", err.Error())
		}
		for _, conn := range connections {

			input := &apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: aws.String(conn),
				Data:         []byte(resp),
			}

			_, err = apigateway.PostToConnection(input)
			if err != nil {
				log.Println("ERROR while sending message to a client", err.Error())
			}
		}
	}
	return nil

}
