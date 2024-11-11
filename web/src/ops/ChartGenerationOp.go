package ops

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"web/src/model"
)

type ChartGenerationOp struct {
	dataFile model.DataFile
}

func NewChartGenerationOp(dataFile model.DataFile) *ChartGenerationOp {
	return &ChartGenerationOp{dataFile}
}

func (op *ChartGenerationOp) Retries() int {
	return 2
}

func (op *ChartGenerationOp) Run(input interface{}) (interface{}, error) {
	code, ok := input.(string)
	if !ok {
		return nil, errors.New("invalid input type for ChartGenerationOp")
	}

	log.Println("Python code:\n", code)

	// Run the python code to generate a chart
	chart, err := executePythonCode(code, op.dataFile)
	if err != nil {
		return nil, err
	}
	return chart.Chart, nil
}

const pythonAPIURL = "http://localhost:7000/generate-chart/"

func executePythonCode(code string, dataFile model.DataFile) (model.PythonCodeResponse, error) {
	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the code field to the form
	codePart, err := writer.CreateFormField("code")
	if err != nil {
		return model.PythonCodeResponse{}, fmt.Errorf("error creating form field for code: %v", err)
	}
	_, err = codePart.Write([]byte(code))
	if err != nil {
		return model.PythonCodeResponse{}, fmt.Errorf("error writing code to form field: %v", err)
	}

	// Add the file field using the file content in memory
	filePart, err := writer.CreateFormFile("file", "data."+dataFile.Ext)
	if err != nil {
		return model.PythonCodeResponse{}, fmt.Errorf("error creating form file for upload: %v", err)
	}
	_, err = io.Copy(filePart, bytes.NewReader(dataFile.Data))
	if err != nil {
		return model.PythonCodeResponse{}, fmt.Errorf("error copying file data: %v", err)
	}

	// Close the writer to finalize the form
	err = writer.Close()
	if err != nil {
		return model.PythonCodeResponse{}, fmt.Errorf("error closing writer: %v", err)
	}

	// Create the HTTP request with the multipart form data
	req, err := http.NewRequest("POST", pythonAPIURL, &requestBody)
	if err != nil {
		return model.PythonCodeResponse{}, fmt.Errorf("error creating HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.PythonCodeResponse{}, fmt.Errorf("error making request to Python API: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("Failed to close response body", err)
		}
	}(resp.Body)

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.PythonCodeResponse{}, fmt.Errorf("error reading response body: %v", err)
	}
	if resp.StatusCode > 399 {
		return model.PythonCodeResponse{}, fmt.Errorf("error from the python environment: %s", body)
	}

	// Parse the response
	var pythonResponse model.PythonCodeResponse
	if err := json.Unmarshal(body, &pythonResponse); err != nil {
		return model.PythonCodeResponse{}, fmt.Errorf("error decoding JSON response: %v", err)
	}

	return pythonResponse, nil
}
