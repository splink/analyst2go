package model

import (
	"fmt"
	"strings"
)

type DataFile struct {
	Headers   []string
	FirstRows [][]string
	Data      []byte
	Ext       string
}

type AnalysisOption struct {
	Name        string   `json:"name"`
	ChartType   string   `json:"chart_type"`
	Description string   `json:"description"`
	Columns     []string `json:"columns"`
}

type CodeResponse struct {
	Status string `json:"status"`
	Code   string `json:"code,omitempty"`
}

type AnalysisOptions struct {
	AnalysisOptions []AnalysisOption `json:"analysis_options"`
}

// HeadersString returns a plain text representation of Headers
func (df *DataFile) HeadersString() string {
	var formattedHeaders []string
	for _, header := range df.Headers {
		// Strip whitespace, convert to lowercase, and replace spaces with underscores
		formattedHeader := strings.ToLower(strings.TrimSpace(header))
		formattedHeader = strings.ReplaceAll(formattedHeader, " ", "_")
		formattedHeaders = append(formattedHeaders, formattedHeader)
	}
	return fmt.Sprintf("Headers: %s", strings.Join(formattedHeaders, ", "))
}

// FirstRowsString returns a plain text representation of the first 5 rows
func (df *DataFile) FirstRowsString() string {
	var formattedRows []string
	for i, row := range df.FirstRows {
		// Limit output to the first 5 rows
		if i >= 5 {
			break
		}
		formattedRows = append(formattedRows, fmt.Sprintf("Row %d: %s", i+1, strings.Join(row, ", ")))
	}
	return fmt.Sprintf("First 5 Rows:\n%s", strings.Join(formattedRows, "\n"))
}

// ChatGPTResponse represents the JSON structure returned by the ChatGPT API.
type ChatGPTResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// PythonCodeRequest represents the JSON payload for the Python code request.
type PythonCodeRequest struct {
	Code     string `json:"code"`
	Data     []byte `json:"data"`
	FileName string `json:"file_name"`
}

// PythonCodeResponse represents the JSON response from the Python API.
type PythonCodeResponse struct {
	Chart string `json:"chart"`
}
