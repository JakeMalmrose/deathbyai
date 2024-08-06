package types

type Game struct {
	GameId        string            `json:"gameId" dynamodbav:"GameId"`
	CreatorId     string            `json:"creatorId" dynamodbav:"CreatorId"`
	Status        string            `json:"status" dynamodbav:"Status"`
	Players       []string          `json:"players" dynamodbav:"Players"`
	CurrentPrompt string            `json:"currentPrompt" dynamodbav:"CurrentPrompt"`
	Responses     map[string]string `json:"responses" dynamodbav:"Responses"`
	Results       map[string]bool   `json:"results" dynamodbav:"Results"`
	CreatedAt     int64             `json:"createdAt" dynamodbav:"CreatedAt"`
	UpdatedAt     int64             `json:"updatedAt" dynamodbav:"UpdatedAt"`
	MaxPlayers    int               `json:"maxPlayers" dynamodbav:"MaxPlayers"`
}

// You can add other shared types here as needed