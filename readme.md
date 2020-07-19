# lambda-websockets

This app requires a valid AWS account, the AWS CLI installed and configured to work with the valid account,
an available MongoDB instance on the Cloud with the appropriate credentials.

## Build the websocket server

Run the command `env GOOS=linux go build -ldflags="-s -w" -o ./bin/handleRequest ./handleRequest`

## Deploy the Lambda function

To deploy the WebSocket server on AWS run the command
`sls deploy --mongo_uri "mongodb+srv://usr:pwd@my-cluster.mongodb.net/my_db_name?retryWrites=true&w=majority" --mongo_database my_db_name`

## How to test manually the WebSocets server

Install wscat with the command `npm install -g wscat`
Run the command `wscat -c wss://may-lambda-endpoint` where `my-lambda-endpoint` is the endpoint creating during the deployment phase, inclusi of the stage, i.e. something like `abc123xyz.execute-api.us-east-1.amazonaws.com/dev`.
This command opens a prompt where you can enter any message string to be sent to the server.
