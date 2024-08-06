package main

import (
	"context"
	"encoding/json"
	"os"
	"time"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

type Game struct {
	GameId       string            `json:"gameId" dynamodbav:"GameId"`
	CreatorId    string            `json:"creatorId" dynamodbav:"CreatorId"`
	Status       string            `json:"status" dynamodbav:"Status"`
	Players      []string          `json:"players" dynamodbav:"Players"`
	CurrentPrompt string           `json:"currentPrompt" dynamodbav:"CurrentPrompt"`
	Responses    map[string]string `json:"responses" dynamodbav:"Responses"`
	Results      map[string]bool   `json:"results" dynamodbav:"Results"`
	CreatedAt    int64             `json:"createdAt" dynamodbav:"CreatedAt"`
	UpdatedAt    int64             `json:"updatedAt" dynamodbav:"UpdatedAt"`
	MaxPlayers   int               `json:"maxPlayers" dynamodbav:"MaxPlayers"`
}

var dynaClient *dynamodb.DynamoDB

func init() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	dynaClient = dynamodb.New(sess)
}

func createGame(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract creator ID from the request body
	var requestBody struct {
		CreatorId string `json:"creatorId"`
	}
	err := json.Unmarshal([]byte(request.Body), &requestBody)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request body"}, nil
	}

	// Create a new game
	now := time.Now().Unix()
	game := Game{
		GameId:       uuid.New().String(),
		CreatorId:    requestBody.CreatorId,
		Status:       "waiting",
		Players:      []string{requestBody.CreatorId},
		CurrentPrompt: "",
		Responses:    make(map[string]string),
		Results:      make(map[string]bool),
		CreatedAt:    now,
		UpdatedAt:    now,
		MaxPlayers:   8,
	}

	// Convert game to DynamoDB AttributeValue
	av, err := dynamodbattribute.MarshalMap(game)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Error marshalling game"}, nil
	}

	// Store the game in DynamoDB
	tableName := os.Getenv("GAMES_TABLE")
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = dynaClient.PutItem(input)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body: fmt.Sprintf("Error storing game in DynamoDB: %v", err),
		}, nil
	}

	// Return the created game
	responseBody, err := json.Marshal(game)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Error marshalling response"}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

func main() {
	lambda.Start(createGame)
}