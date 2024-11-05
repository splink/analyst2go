package main

import (
	"bytes"
	"fmt"
	"context"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
)

const maxRetries = 5
const apiKey = "YOUR_OPENAI_API_KEY"
const pythonAPIURL = "http://localhost:7000/generate-chart/"

// ChatGPTResponse represents the JSON structure returned by the ChatGPT API.
type ChatGPTResponse struct {
	Status  string                   `json:"status"`
	Data    []map[string]interface{} `json:"data,omitempty"`
	Message string                   `json:"message,omitempty"`
}

// PythonCodeRequest represents the JSON payload for the Python code request.
type PythonCodeRequest struct {
	Code string `json:"code"`
}

// PythonCodeResponse represents the JSON response from the Python API.
type PythonCodeResponse struct {
	Chart  string `json:"chart"`
	Format string `json:"format"`
}

func main() {
	r := gin.Default()

	// Load HTML templates from the templates directory
	r.LoadHTMLGlob("templates/*")

	// Route to display the HTML page with embedded Plotly chart
	r.GET("/", func(c *gin.Context) {
		// Define the Python code as a string
		pythonCode := `
import plotly.express as px

# Sample data
data = {
    'x': [1, 2, 3, 4, 5],
    'y': [10, 15, 13, 17, 20]
}

# Generate Plotly figure
fig = px.line(data, x="x", y="y", title="Sample Plotly Chart", template="simple_white")
output = fig
`

		// Create the JSON payload for Python API
		payload := PythonCodeRequest{Code: pythonCode}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Fatalf("Error encoding JSON: %v", err)
		}

		// Send the request to the Python API
		resp, err := http.Post(pythonAPIURL, "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			log.Fatalf("Error making request to Python API: %v", err)
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading response body: %v", err)
		}

		// Unmarshal the response
		var pythonResponse PythonCodeResponse
		if err := json.Unmarshal(body, &pythonResponse); err != nil {
			log.Fatalf("Error decoding JSON response: %v", err)
		}

		// Pass the Plotly JSON data to the template
		c.HTML(http.StatusOK, "index.html", gin.H{
			"PlotlyJSON": template.JS(pythonResponse.Chart), // Use template.JS to avoid escaping
		})
	})

	// New route to handle image upload and call ChatGPT for data extraction
	r.POST("/upload", func(c *gin.Context) {
		file, _, err := c.Request.FormFile("screenshot")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read file"})
			return
		}
		defer file.Close()

		// Read file into a buffer
		imageData, err := ioutil.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image data"})
			return
		}

		// Base64 encode the image for JSON transmission
		base64Image := base64.StdEncoding.EncodeToString(imageData)

		// Create OpenAI client
		client := openai.NewClient(apiKey)
		ctx := context.Background()

		// Configure the chat prompt
		req := openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "system",
					Content: "You are a data extraction tool specialized in reading structured tables from images. " +
						"Please identify and extract tabular data, and format it as a JSON array of rows, where each row contains key-value pairs. " +
						"If headers are present, use them as keys. If no tabular data exists, respond with an error status and an explanatory message.",
				},
				{
					Role:    "user",
					Content: fmt.Sprintf("Extract data from this image: data:image/png;base64,%s", base64Image),
				},
			},
		}

		// Retry with exponential back-off on failure
		var chatResponse ChatGPTResponse
		success := false
		for i := 0; i < maxRetries; i++ {
			var resp openai.ChatCompletionResponse
			resp, err = client.CreateChatCompletion(ctx, req)
			if err == nil && len(resp.Choices) > 0 {
				// Attempt to parse the response content as JSON
				err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &chatResponse)
				if err == nil && chatResponse.Status == "ok" {
					success = true
					break
				}

				// Check for an error status response indicating no data
				if chatResponse.Status == "error" && chatResponse.Message == "No extractable data. This might not contain usable information." {
					break
				}
			}

			// Exponential back-off
			log.Printf("Request failed: %v. Retrying in %d seconds...", err, int(math.Pow(2, float64(i))))
			time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
		}

		// Check if request succeeded
		if !success {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract data after retries"})
			return
		}

		// Return extracted data as JSON
		c.JSON(http.StatusOK, chatResponse)
	})

	// Run the server on port 8080
	r.Run(":8080")
}

