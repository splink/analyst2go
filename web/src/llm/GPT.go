package llm

import (
	"context"
	"encoding/json"
	"github.com/sashabaranov/go-openai"
	"log"
	"math"
	"strings"
	"time"
	"web/src/model"
	"web/src/util"
)

var maxRetries = 5

func SendToGPTWithRetry(req openai.ChatCompletionRequest) (model.ChatGPTResponse, bool) {
	apiKey := util.Env("OPENAI_API_KEY")
	client := openai.NewClient(apiKey)
	ctx := context.Background()

	var chatResponse model.ChatGPTResponse
	for i := 0; i < maxRetries; i++ {
		resp, err := client.CreateChatCompletion(ctx, req)
		if err == nil && len(resp.Choices) > 0 {
			content := fixJSON(resp.Choices[0].Message.Content)
			err = json.Unmarshal([]byte(content), &chatResponse)
			if err == nil && (chatResponse.Status == "ok" || chatResponse.Status == "error") {
				return chatResponse, chatResponse.Status == "ok"
			}
		}

		log.Printf("Request failed: %v. Retrying in %d seconds...", err, int(math.Pow(2, float64(i))))
		time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
	}

	return chatResponse, false
}

func fixJSON(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	startPos := strings.IndexRune(s, '{')
	endPos := strings.LastIndex(s, "}")

	if startPos == -1 || endPos == -1 || endPos < startPos {
		return ""
	}
	return s[startPos : endPos+1]
}
