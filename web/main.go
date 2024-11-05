package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

		// Create the JSON payload
		payload := PythonCodeRequest{Code: pythonCode}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Fatalf("Error encoding JSON: %v", err)
		}

		// Send the request to the Python API
		resp, err := http.Post("http://localhost:7000/generate-chart/", "application/json", bytes.NewBuffer(payloadBytes))
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

	// Run the server on port 8080
	r.Run(":8080")
}

