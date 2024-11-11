package llm

import (
	"context"
	"encoding/json"
	"github.com/sashabaranov/go-openai"
	"log"
	"math"
	"strings"
	"time"
	"web/src/util"
)

var maxRetries = 5

func SendToGPTWithRetry(req openai.ChatCompletionRequest) (string, bool) {
	apiKey := util.Env("OPENAI_API_KEY")
	client := openai.NewClient(apiKey)
	ctx := context.Background()

	for i := 0; i < maxRetries; i++ {
		resp, err := client.CreateChatCompletion(ctx, req)
		if err == nil && len(resp.Choices) > 0 {
			content := fixJSON(resp.Choices[0].Message.Content)
			isValidJson := json.Valid([]byte(content))
			if isValidJson {
				return content, true
			}
			// JSON is invalid, so log and retry
			log.Printf("Received invalid JSON. Retrying... (Attempt %d of %d)", i+1, maxRetries)
		} else {
			// Request failed for another reason, log and retry
			log.Printf("Request failed: %v. Retrying in %d seconds... (Attempt %d of %d)", err, int(math.Pow(2, float64(i))), i+1, maxRetries)
		}

		time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
	}

	return "", false
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
