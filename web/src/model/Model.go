package model

// ChatGPTResponse represents the JSON structure returned by the ChatGPT API.
type ChatGPTResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// PythonCodeRequest represents the JSON payload for the Python code request.
type PythonCodeRequest struct {
	Code string `json:"code"`
}

// PythonCodeResponse represents the JSON response from the Python API.
type PythonCodeResponse struct {
	Chart string `json:"chart"`
}
