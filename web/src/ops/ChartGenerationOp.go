package ops

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"web/src/model"
)

type ChartGenerationOp struct{}

func (op *ChartGenerationOp) Retries() int {
	return 2
}

func (op *ChartGenerationOp) Run(input interface{}) (interface{}, error) {
	pythonCodeToRun, ok := input.(string)
	if !ok {
		return nil, errors.New("invalid input type for ChartGenerationOp")
	}

	// Run the python code to generate a chart
	chart, err := executePythonCode(model.PythonCodeRequest{Code: pythonCodeToRun})
	if err != nil {
		return nil, err
	}
	return chart.Chart, nil
}

const pythonAPIURL = "http://localhost:7000/generate-chart/"

// executePythonCode sends Python code to a Python API for execution and returns the response or an error.
func executePythonCode(request model.PythonCodeRequest) (model.PythonCodeResponse, error) {
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return model.PythonCodeResponse{}, fmt.Errorf("error encoding JSON: %v", err)
	}

	resp, err := http.Post(pythonAPIURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return model.PythonCodeResponse{}, fmt.Errorf("error making request to Python API: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("error closing response body: ", err)
		}
	}(resp.Body)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return model.PythonCodeResponse{}, fmt.Errorf("error reading response body: %v", err)
	}
	if resp.StatusCode > 399 {
		return model.PythonCodeResponse{}, fmt.Errorf("error from the python environment: %s", body)
	}

	var pythonResponse model.PythonCodeResponse
	if err := json.Unmarshal(body, &pythonResponse); err != nil {
		return model.PythonCodeResponse{}, fmt.Errorf("error decoding JSON response: %v", err)
	}

	return pythonResponse, nil
}
