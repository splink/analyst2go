package main

import (
	"github.com/gin-gonic/gin"
	"html/template"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"web/src/ops"
	"web/src/util"
)

const maxRetries = 8

func main() {
	util.LoadEnvVars()

	baseURL := util.Env("BASE_URL")
	if strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL[:len(baseURL)-1]
	}

	r := gin.Default()

	// Load HTML templates from the templates directory
	r.LoadHTMLGlob("templates/*")

	// Route to display the HTML page with embedded Plotly chart
	r.GET("/", func(c *gin.Context) {
		// Define the Python code as a string
		pythonCode := `
import plotly.express as px
import pandas as pd

# Sample data
data = {
    'Year': [2018, 2019, 2020, 2021, 2022],
    'Revenue ($M)': [10, 15, 13, 17, 20],
    'Company': ["Alpha Corp", "Beta Inc", "Gamma LLC", "Delta Co", "Epsilon Ltd"]
}

# Create DataFrame
df = pd.DataFrame(data)

# Generate Plotly figure
fig = px.line(
    df,
    x="Year",
    y="Revenue ($M)",
    title="Company Revenue Over Time",
    labels={"Year": "Year", "Revenue ($M)": "Revenue in Millions (USD)"},
    template="simple_white"
)

# Enhance trace for readability
fig.update_traces(mode="lines+markers", hovertemplate="Company: %{text}<br>Revenue: %{y:.1f}M<br>Year: %{x}")
fig.update_traces(text=df['Company'])

output = fig

`
		op := ops.ChartGenerationOp{}
		result, err := op.Run(pythonCode)
		if err != nil {
			log.Println("Failed to generate chart:", err)
			c.String(http.StatusInternalServerError, "Oops! Something went wrong.")
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"PlotlyJSON": template.JS(result.(string)),
		})
	})

	r.POST("/uploadImage", handleImage)
	r.POST("/uploadFile", handleFile)

	// Run the server on port 8080
	err := r.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
		return
	}
}

// handleFile handles Excel and CSV file uploads
func handleFile(c *gin.Context) {
	now := time.Now()
	log.Println("File upload received")
	// Get file from the request
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "File upload error: %v", err)
		return
	}

	// Validate file extension
	allowedExtensions := map[string]bool{".csv": true, ".xls": true, ".xlsx": true}
	ext := filepath.Ext(file.Filename)
	if !allowedExtensions[ext] {
		c.String(http.StatusBadRequest, "Invalid file type. Only CSV and Excel files are allowed.")
		return
	}

	// Process the uploaded file
	fileData, err := readFileData(file)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to read file data: %v", err)
		return
	}

	elapsed := time.Since(now)
	log.Println(elapsed.Seconds(), "File data processed", string(fileData))

	pipeline := ops.NewPipeline(
		&ops.DataAnalysisOp{},
		&ops.ChartGenerationOp{},
	)

	chart, err := pipeline.Execute(string(fileData))
	if err != nil {
		log.Println("Pipeline execution failed:", err)
		c.String(http.StatusInternalServerError, "Failed to process image data")
		return
	}

	extractedData, _ := pipeline.GetResult(0)
	analysisCode, _ := pipeline.GetResult(1)

	c.HTML(http.StatusOK, "index.html", gin.H{
		"PlotlyJSON": template.JS(chart.(string)),
		"Data":       extractedData.(string),
		"Code":       analysisCode.(string),
	})
}

// readFileData reads the uploaded file and returns its contents as a byte slice
func readFileData(file *multipart.FileHeader) ([]byte, error) {
	uploadedFile, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer func(uploadedFile multipart.File) {
		err := uploadedFile.Close()
		if err != nil {
			log.Println("Failed to close file", err)
		}
	}(uploadedFile)

	fileData := make([]byte, file.Size)
	_, err = uploadedFile.Read(fileData)
	if err != nil {
		return nil, err
	}

	return fileData, nil
}

func handleImage(c *gin.Context) {
	imageData, err := readImage(c)
	if err != nil {
		log.Println("Failed to read image data:", err)
		c.String(http.StatusInternalServerError, "Failed to read image data")
		return
	}

	pipeline := ops.NewPipeline(
		&ops.ImageDataExtractionOp{},
		&ops.DataAnalysisOp{},
		&ops.ChartGenerationOp{},
	)

	chart, err := pipeline.Execute(imageData)
	if err != nil {
		log.Println("Pipeline execution failed:", err)
		c.String(http.StatusInternalServerError, "Failed to process image data")
		return
	}

	extractedData, _ := pipeline.GetResult(0)
	analysisCode, _ := pipeline.GetResult(1)

	c.HTML(http.StatusOK, "index.html", gin.H{
		"PlotlyJSON": template.JS(chart.(string)),
		"Data":       extractedData.(string),
		"Code":       analysisCode.(string),
	})
}

func readImage(c *gin.Context) ([]byte, error) {
	file, _, err := c.Request.FormFile("screenshot")
	if err != nil {
		return nil, err
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			log.Println("Failed to close file")
		}
	}(file)

	return ioutil.ReadAll(file)
}
