package ops

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"log"
	"web/src/llm"
)

type ImageDataExtractionOp struct{}

func (op *ImageDataExtractionOp) Retries() int {
	return 3
}

func (op *ImageDataExtractionOp) Run(input interface{}) (interface{}, error) {
	imageData, ok := input.([]byte)
	if !ok {
		return nil, errors.New("invalid input type for ImageDataExtractionOp")
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)

	request := createImageExtractionRequest(base64Image)
	response, success := llm.SendToGPTWithRetry(request)
	if !success {
		log.Println("Failed to extract data from image")
		return nil, errors.New("failed to extract data from image")
	}

	return response, nil
}

func createImageExtractionRequest(base64Image string) openai.ChatCompletionRequest {
	imagePart := openai.ChatMessagePart{
		Type: openai.ChatMessagePartType("image_url"),
		ImageURL: &openai.ChatMessageImageURL{
			URL:    fmt.Sprintf("data:image/png;base64,%s", base64Image),
			Detail: "high",
		},
	}

	textPart := openai.ChatMessagePart{
		Type: openai.ChatMessagePartType("text"),
		Text: "Create a CSV representation of the data in the image",
	}

	return openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "system",
				Content: `You are a data extraction tool specialized in reading structured tables from images. 
Respond with a valid json document following this structure:
{"status": "ok", "message": "message"} 
Set the status to either "ok" or "error" and provide result in the message field.`,
			},
			{
				Role:         "user",
				MultiContent: []openai.ChatMessagePart{textPart, imagePart},
			},
		},
	}
}
