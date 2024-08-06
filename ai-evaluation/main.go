package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/bedrockruntime"

	"deathbyai/types"
)



type EvaluationRequest struct {
	GameId   string `json:"gameId"`
	PlayerId string `json:"playerId"`
	Response string `json:"response"`
}

type EvaluationResponse struct {
	Survived bool   `json:"survived"`
	Explanation string `json:"explanation"`
}

var dynaClient *dynamodb.DynamoDB
var bedrockClient *bedrockruntime.BedrockRuntime

func init() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	dynaClient = dynamodb.New(sess)
	bedrockClient = bedrockruntime.New(sess)
}

func evaluateResponse(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var evalRequest EvaluationRequest
	err := json.Unmarshal([]byte(request.Body), &evalRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request body"}, nil
	}

	// Retrieve game from DynamoDB
	game, err := getGame(evalRequest.GameId)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Error retrieving game"}, nil
	}

	// Construct prompt for AI
	prompt := constructPrompt(game.CurrentPrompt, evalRequest.Response)

	// Call AI for evaluation
	survived, explanation, err := callAI(prompt)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Error calling AI"}, nil
	}

	// Update game in DynamoDB with result
	game.Responses[evalRequest.PlayerId] = evalRequest.Response
	game.Results[evalRequest.PlayerId] = survived
	err = updateGame(game)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Error updating game"}, nil
	}

	// Prepare and return response
	evalResponse := EvaluationResponse{
		Survived: survived,
		Explanation: explanation,
	}
	responseBody, _ := json.Marshal(evalResponse)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{"Content-Type": "application/json"},
		Body: string(responseBody),
	}, nil
}

func getGame(gameId string) (*types.Game, error) {
	result, err := dynaClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("GAMES_TABLE")),
		Key: map[string]*dynamodb.AttributeValue{
			"GameId": {S: aws.String(gameId)},
		},
	})
	if err != nil {
		return nil, err
	}

	game := &Game{}
	err = dynamodbattribute.UnmarshalMap(result.Item, game)
	return game, err
}

func updateGame(game *types.Game) error {
	av, err := dynamodbattribute.MarshalMap(game)
	if err != nil {
		return err
	}

	_, err = dynaClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("GAMES_TABLE")),
		Item:      av,
	})
	return err
}

func constructPrompt(scenario, response string) string {
	return fmt.Sprintf(`Given the following scenario:
"%s"

A player responded with:
"%s"

Determine if the player would survive or die based on their response. Provide a brief explanation of the outcome.

Output your response in the following format:
Survived: [true/false]
Explanation: [Your explanation here]`, scenario, response)
}

func callAI(prompt string) (bool, string, error) {
	// This is a placeholder for the actual AI call
	// You would replace this with a call to Amazon Bedrock
	// For now, we'll return a dummy response
	return true, "The player's quick thinking and resourcefulness led to their survival.", nil
}

func main() {
	lambda.Start(evaluateResponse)
}