package ops

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"log"
	"web/src/llm"
	"web/src/model"
)

type DataAnalysisOp struct{}

func (op *DataAnalysisOp) Retries() int {
	return 3
}

func (op *DataAnalysisOp) Run(input interface{}) (interface{}, error) {
	data, ok := input.(model.DataFile)
	if !ok {
		return nil, errors.New("invalid input type for DataAnalysisOp")
	}

	request := createDataAnalysisRequest(data)
	response, success := llm.SendToGPTWithRetry(request)
	if !success {
		log.Println("Failed to analyze extracted data")
		return nil, errors.New("failed to analyze extracted data")
	}

	var codeResponse model.CodeResponse
	err := json.Unmarshal([]byte(response), &codeResponse)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return nil, err
	}

	return codeResponse.Code, nil
}

// createDataAnalysisRequest constructs a request payload for analyzing extracted data to generate Python Pandas code
func createDataAnalysisRequest(data model.DataFile) openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "system",
				Content: `You are an AI assistant responsible for generating Python code to perform a data analysis and produce a Plotly chart as output. 
A DataFrame named "df"" containing the uploaded data is already in scope. This is the code that is run before your code:

	if file_extension == "csv":
        df = pd.read_csv(io.BytesIO(file_content))
    elif file_extension in ["xls", "xlsx"]:
        df = pd.read_excel(io.BytesIO(file_content))

	df.dropna(how="all", inplace=True)
	df.dropna(axis=1, how="all", inplace=True)
	df.columns = df.columns.str.strip().str.lower().str.replace(' ', '_')
	df.reset_index(drop=True, inplace=True)
	exec_globals = {"pd": pd, "np": np, "px": px, "go": go, "df": df, "output": None}
	exec(code, exec_globals)

Your code must be self-contained, reliable, and compatible with the following Python libraries:

numpy==1.23.5
pandas==1.5.3
plotly==5.11.0
scikit-learn==1.2.2
Please follow these instructions carefully:

1. Data Handling:
Assume df is fully loaded and contains the data. Use df for all analysis and charting, without redefining or reloading the data.
Note that all column names are normalized. A column name is always in lower case and has an underscore instead of spaces. e.g. "Column Name" -> "column_name"


2. Code Structure and Output:
Follow the provided structure in the example. Ensure all code is encapsulated in a single, self-contained code block. Do not write comments. 
Make sure to produce Python code that can be executed without errors - Python requires correct indentation and syntax.
Use plotly.express to create a chart based on the data in df, and assign the variable fig = px.bar(...) to the variable output, so the code ends with output = fig.


3. Chart Quality:
Select a chart type and variables that best illustrate the relationships or trends in the data. Use line, bar, or scatter charts 
as appropriate for time series, categorical, or numerical data.

Use descriptive axis labels, a clear title, and add tooltips to enhance readability, allowing for intuitive exploration of the chart.
Ensure no overlapping text or clutter in the chart.


4. Error Handling and Validation:
Ensure the code executes correctly without requiring further adjustments to the data loading or structure.


5. Output Format:

Respond in valid JSON format:
{ "status": "ok",  "code": "<code>" }
Set "status" to "ok" if the code is correct, or "error" if corrections are needed. Place the Python code inside the "message" field.


Example Code Structure:

import plotly.express as px
import pandas as pd
import numpy as np

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
`,
			},
			{
				Role: "user",
				Content: fmt.Sprintf(`The shape of the data:
%s
%s

Analysis 
Check the available data and it's structure. Check the types of the columns and the relationships between them. 
Evaluate the data to determine the most impactful analysis to perform and select the most suitable chart type.
For instance, when the data contains time series data, a line chart may be appropriate.
Then write code to perform the analysis and generate a Plotly chart based on the data.

Respond with a JSON object in the following format, where you insert the Python code in place of "<code>":
{
  "status": "ok",
  "code": "<code>"
}`, data.HeadersString(), data.FirstRowsString()),
			},
		},
	}
}
