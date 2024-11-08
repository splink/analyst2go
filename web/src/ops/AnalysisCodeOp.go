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
				Content: `You are an AI assistant responsible for generating Python code to analyze data and produce a Plotly chart as output. The code must be self-contained, reliable, and compatible with the following Python libraries:
- numpy==1.23.5
- pandas==1.5.3
- plotly==5.11.0
- scikit-learn==1.2.2

Please follow these instructions carefully:

1. Data Handling:
Include the complete dataset within the response as part of a self-contained code snippet.
Ensure the data is loaded into a pandas.DataFrame in a format suitable for analysis with the pandas library.
In case the provide dataset contains missing values handle them by imputation and substitute them with the mean value of the respective column.
For instance, make sure that all arrays are of the same length.

2. Code Structure and Output:
The code should match the structure of the example provided below.
Use the plotly.express library to create a single chart and assign it to the variable output.
Ensure that the chart is labeled clearly, with well-placed hover information, readable labels, and no overlapping text.

3. Chart Quality:
The generated chart should be visually appealing, easy to interpret, and provide clear insights.
Use descriptive labels for both axes and title to enhance readability.
Ensure that hover tooltips are informative, providing relevant details about data points.

4. Example code:
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



5. Error Handling:
Perform a quality check to ensure the generated code is executable.
If the data format is incompatible with pd.DataFrame, adjust it accordingly within the code.

6. Output Format:
Ensure that your response always is a valid json document with the following structure, where message contains the Python code:
{"status": "ok", "message": "message"}
Set the status to either "ok" or "error" and provide the code in the message field.`,
			},
			{
				Role: "user",
				Content: fmt.Sprintf(`Given the following data:\n%s\n
Generate the code to perform the most insightful and relevant analysis based on the available data. 
Create a chart that visualizes the analysis in a powerful way that is simple to understand.`, extractedData),
			},
		},
	}
}
