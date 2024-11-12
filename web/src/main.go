package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"web/src/db"
	"web/src/model"
	"web/src/ops"
	"web/src/service"
	"web/src/util"
)

func main() {
	util.LoadEnvVars()
	db.Init()

	baseURL := util.Env("BASE_URL")
	if strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL[:len(baseURL)-1]
	}

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", index)
	//r.POST("/uploadImage", handleImage)
	r.POST("/uploadFile", handleFile)

	err := r.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
		return
	}
}

func index(c *gin.Context) {
	// Define the Python code as a string
	pythonCode := `
import plotly.express as px

# Melt DataFrame to have subjects in one column and scores in another
df_melted = df.melt(id_vars=["studentid", "first_name", "last_name", "attendance"],
                    value_vars=["math", "phonics", "science"],
                    var_name="subject", value_name="score")

# Create a bar chart showing scores for each subject by student
fig = px.bar(
    df_melted,
    x="first_name",
    y="score",
    color="subject",
    title="Student Scores by Subject",
    labels={"first_name": "Student", "score": "Score"},
    template="simple_white"
)

output = fig
`

	data, _ := ioutil.ReadFile("/home/maxmc/Downloads/student_data.csv")

	op := ops.NewChartGenerationOp(model.DataFile{
		Headers:   []string{"studentid", "first_name", "last_name", "attendance", "math", "phonics", "science"},
		FirstRows: [][]string{{"1", "John", "Doe", "Present", "90", "85", "88"}, {"2", "Jane", "Smith", "Absent", "85", "90", "92"}},
		Data:      data,
		Ext:       "csv",
	})

	result, err := op.Run(pythonCode)
	if err != nil {
		log.Println("Failed to generate chart:", err)
		c.String(http.StatusInternalServerError, "Oops! Something went wrong.")
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"PlotlyJSON": template.JS(result.(string)),
	})
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

	// Process the uploaded file
	dataFile, err := readFileData(file)
	if err != nil {
		c.String(http.StatusInternalServerError, "There is a problem with the file data: %v", err)
		return
	}

	insightID, err := service.CreateInsight(1)
	if err != nil {
		log.Println("Failed to create insight:", err)
		return
	}
	err = service.SaveInsightData(insightID, dataFile)
	if err != nil {
		log.Println("Failed to save insight data:", err)
		return
	}

	c.String(http.StatusOK, "File uploaded successfully")
	return

	//TODO testing options
	op := ops.DataAnalysisOptionsOp{}
	options, err := op.Run(dataFile)
	if err != nil {
		log.Println("Failed to generate analysis options:", err)
		return
	}

	log.Println("option", options)

	elapsed := time.Since(now)
	log.Printf("%.2f %s File data processed", elapsed.Seconds(), dataFile.Ext)
	pipeline := ops.NewPipeline(
		&ops.DataAnalysisOp{},
		ops.NewChartGenerationOp(dataFile),
	)

	chart, err := pipeline.Execute(dataFile)
	if err != nil {
		log.Println("Pipeline execution failed:", err)
		c.String(http.StatusInternalServerError, "Failed to process pipeline")
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

const maxRows = 100
const maxFileSize = 10 * 1024 * 1024 // 10MB limit for file size

func readFileData(file *multipart.FileHeader) (model.DataFile, error) {
	allowedExtensions := map[string]bool{".csv": true, ".xls": true, ".xlsx": true}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		return model.DataFile{}, fmt.Errorf("invalid file type. Only CSV and Excel files are allowed")
	}

	// Open the file
	multipartFile, err := file.Open()
	if err != nil {
		return model.DataFile{}, err
	}
	defer func() {
		if closeErr := multipartFile.Close(); closeErr != nil {
			log.Println("Failed to close file:", closeErr)
		}
	}()

	if file.Size > maxFileSize {
		return model.DataFile{}, fmt.Errorf("file size exceeds the limit of %d bytes", maxFileSize)
	}
	fileData := make([]byte, file.Size)
	_, err = io.ReadFull(multipartFile, fileData)
	if err != nil {
		return model.DataFile{}, fmt.Errorf("failed to read file data: %v", err)
	}

	// Reset the file reader to use in parsing
	reader := bytes.NewReader(fileData)

	var headers []string
	var rows [][]string
	rowCount := 0

	if ext == ".csv" {
		// Process CSV file
		reader := csv.NewReader(reader)
		headers, err = reader.Read()
		if err != nil {
			return model.DataFile{}, fmt.Errorf("failed to read CSV headers: %v", err)
		}

		for {
			row, err := reader.Read()
			if err != nil {
				break // end of file
			}
			rows = append(rows, row)
			rowCount++
			if rowCount >= maxRows {
				break
			}
		}

		if rowCount < 1 {
			return model.DataFile{}, fmt.Errorf("file must contain at least two rows")
		}

	} else {
		// Process Excel file
		excelFile, err := excelize.OpenReader(reader)
		if err != nil {
			return model.DataFile{}, fmt.Errorf("failed to read Excel file: %v", err)
		}

		sheetName := excelFile.GetSheetName(1)
		excelRows, err := excelFile.GetRows(sheetName)
		if err != nil || len(excelRows) == 0 {
			return model.DataFile{}, fmt.Errorf("failed to read rows from Excel sheet")
		}

		headers = excelRows[0]
		rows = excelRows[1:min(maxRows+1, len(excelRows))]
		rowCount = len(rows)
	}

	if len(headers) < 2 {
		return model.DataFile{}, fmt.Errorf("file must contain at least two columns")
	}

	// Validate each header
	for _, header := range headers {
		header = strings.TrimSpace(header)
		if header == "" {
			return model.DataFile{}, fmt.Errorf("headers must not be empty")
		}
		if isNumeric(header) {
			return model.DataFile{}, fmt.Errorf("header must be a non-numeric string: %s", header)
		}
		if !isMeaningfulHeader(header) {
			return model.DataFile{}, fmt.Errorf("header must contain at least one letter: %s", header)
		}
		if !isValidLength(header) {
			return model.DataFile{}, fmt.Errorf("header length must be between 2 and 250 characters: %s", header)
		}
	}

	return model.DataFile{
		Headers:   headers,
		FirstRows: rows,
		Ext:       ext,
		Data:      fileData,
	}, nil
}

// Helper function to check if a string is numeric
func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// Helper function to check if the header has alphabetic characters
func isMeaningfulHeader(header string) bool {
	match, _ := regexp.MatchString(`[a-zA-Z]`, header)
	return match
}

// Helper function to check the length of the header
func isValidLength(header string) bool {
	return len(header) > 1 && len(header) <= 250
}

// readFileData reads the uploaded file and returns its contents as a byte slice
func readFileData2(file *multipart.FileHeader) ([]byte, error) {
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

/*
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
*/
