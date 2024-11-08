package ops

import (
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"log"
	"web/src/llm"
)

type DataAnalysisOp struct{}

func (op *DataAnalysisOp) Retries() int {
	return 3
}

func (op *DataAnalysisOp) Run(input interface{}) (interface{}, error) {
	dataToAnalyze, ok := input.(string)
	if !ok {
		return nil, errors.New("invalid input type for DataAnalysisOp")
	}

	request := createDataAnalysisRequest(dataToAnalyze)
	response, success := llm.SendToGPTWithRetry(request)
	if !success {
		log.Println("Failed to analyze extracted data")
		return nil, errors.New("failed to analyze extracted data")
	}
	if response.Status != "ok" {
		return nil, errors.New("failed to analyze extracted data")
	}

	return response.Message, nil
}

// createDataAnalysisRequest constructs a request payload for analyzing extracted data to generate Python Pandas code
func createDataAnalysisRequest(extractedData string) openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "system",
				Content: `You are an AI assistant responsible for generating Python code to analyze data and produce a Plotly chart as output.
The code must be self-contained, reliable, and compatible with the following Python libraries:

numpy==1.23.5
pandas==1.5.3
plotly==5.11.0
scikit-learn==1.2.2
Please follow these instructions carefully:

1. Data Handling:

Ensure Data Consistency: 
Ensure all columns contain arrays of the same length. Check for and correct any discrepancies in array lengths. If the data contains rows of different lengths or misaligned columns, pad shorter rows with None or NaN and adjust the DataFrame accordingly.

Impute Missing Values: 
If there are missing values, fill them using imputation based on the data type:
* Use the mean for numerical columns.
* Use mode for categorical columns.

Format Correction: 
If the data is in an incorrect format, convert it into a pandas-compatible format (e.g., lists or dictionaries). Ensure that date columns are parsed correctly as datetime where applicable.

Internal Quality Check:
Simulate code execution in your mind to verify that the data is correctly structured and will load without errors. Adjust any issues before proceeding.


2. Code Structure and Output:

Code Structure: 
Follow the provided structure in the example. Ensure all code is encapsulated in a single, self-contained code block.

Chart Creation: 
Use plotly.express to create a single chart, assigning it to the variable output.

Chart Clarity: 
Use descriptive axis labels, a clear title, and add tooltips to enhance readability. Ensure no overlapping text or clutter in the chart.


3. Chart Quality:

Insightful Visualization: 
Select a chart type and variables that best illustrate the relationships or trends in the data. Use line, bar, or scatter charts as appropriate for time series, categorical, or numerical data.

Tooltip Details: 
Ensure hover tooltips provide informative details about each data point, allowing for intuitive exploration of the chart.


4. Error Handling and Validation:

Code Verification: 
Perform a final quality check to ensure the generated code is correct, executable, and free of syntax and runtime errors (e.g., IndentationError, ValueError).

Adjust for Data Quality: 
If the dataset quality is variable (e.g., mixed types or formatting inconsistencies), adjust accordingly to ensure the code can handle such variations robustly.


5. Output Format:

Respond in valid JSON format:
{ "status": "ok",  "message": "<code>" }
Set "status" to "ok" if the code is correct, or "error" if corrections are needed. Place the Python code inside the "message" field.


Example Code Structure:

import plotly.express as px
import pandas as pd
import numpy as np

# Sample Data
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
`,
			},
			{
				Role: "user",
				Content: fmt.Sprintf(`Given the following data:

%s

Data Preparation: 
Fix any formatting issues in the dataset, handle missing values appropriately, and ensure compatibility with pandas.

Chart Creation: 
Write code to perform the most impactful analysis and generate an insightful, easy-to-understand Plotly chart based on the data.

Code Verification: 
Ensure the code is complete and will execute without errors.

Respond with a JSON object in the following format:
{
  "status": "ok",
  "message": "<code>"
}`, extractedData),
			},
		},
	}
}
