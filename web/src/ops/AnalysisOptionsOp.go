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

type DataAnalysisOptionsOp struct{}

func (op *DataAnalysisOptionsOp) Retries() int {
	return 3
}

func (op *DataAnalysisOptionsOp) Run(input interface{}) (interface{}, error) {
	data, ok := input.(model.DataFile)
	if !ok {
		return nil, errors.New("invalid input type for DataAnalysisOp")
	}

	request := createDataAnalysisOptionsRequest(data)
	response, success := llm.SendToGPTWithRetry(request)
	if !success {
		log.Println("Failed to analyze extracted data")
		return nil, errors.New("failed to analyze extracted data")
	}

	var options model.AnalysisOptions

	err := json.Unmarshal([]byte(response), &options)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return nil, errors.New("failed to analyze extracted data")
	}

	return options, nil
}

// createDataAnalysisRequest constructs a request payload for analyzing extracted data to generate Options for Analysis
func createDataAnalysisOptionsRequest(data model.DataFile) openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "system",
				Content: `You are provided with a table of data and a catalog of popular analysis options. Based on the table’s structure and the column types, your task is to identify the 8 most relevant analyses for this dataset. For each analysis option, specify which columns should be used.

Instructions:

1. Examine the data table to identify its structure, column types, and potential analysis methods.
2. Use the catalog below to select 8 suitable analysis options. Prioritize the most impactful and informative analyses for this dataset.
3.  Each analysis option should be provided in the format:
* "name": (The name of the analysis option)
* "description": (A brief description of the analysis)
* "columns": (List of columns relevant to the analysis)
Return your response in JSON format.

Catalog of Analysis Options:

Time Series Analysis

Trend Analysis: Detect overall trends in time-series data.
Seasonality Analysis: Identify seasonal patterns in time-series data.
Rolling Average: Smooth data fluctuations using moving averages.
Time Series Forecasting: Predict future values based on historical data.
Growth Rate Calculation: Assess growth rate over time intervals.
Categorical Data Analysis

Distribution Analysis: Show the distribution of values in a categorical column.
Frequency Count: Count occurrences for each category.
Proportion Analysis: Calculate proportions or percentages within each category.
Top/Bottom Category Analysis: Identify top or bottom categories based on a specific numerical column (e.g., top 5 categories by sales).
Numerical Data Analysis

Summary Statistics: Provide mean, median, standard deviation, min, and max for numerical columns.
Correlation Matrix: Calculate correlations between numerical columns.
Regression Analysis: Explore relationships between numerical columns.
Variance Analysis: Assess variance within a numerical column.
Percentile Distribution: Break down data into percentiles (e.g., 25th, 50th, 75th).
Outlier Detection: Identify outliers in numerical columns.
Mixed Data Analysis (Categorical + Numerical)

Time-Based Grouping with Categories: Show trends of categories over time.
Pivot Table: Summarize data by cross-tabulation of categorical and numerical columns.
Heatmap Analysis: Visualize correlations or intensities between categorical and numerical data.
Comparative Analysis: Compare data across categories or time periods.
Textual Data Analysis (for datasets with textual columns)

Sentiment Analysis: Assess sentiment in textual data.
Keyword Frequency: Count the frequency of keywords or phrases.
Topic Modeling: Identify main topics discussed within textual data.
Text Length Distribution: Analyze distribution of text length across records.
Other Common Analyses

Anomaly Detection: Identify unusual patterns or outliers in numerical or time-based data.
Comparative Analysis: Compare data across categories or time periods.
Top-K Analysis: Identify top values (e.g., top 5 products by sales).
Cohort Analysis: Analyze grouped data over time (e.g., customer cohorts by acquisition month).
Churn Rate Calculation: Calculate the rate of attrition in the dataset (e.g., customer or product churn).
Pareto Analysis: Apply the 80/20 rule to identify key factors contributing most to an outcome.

Available chart types:
scatter, line, bar, pie, histogram, box, violin, density_contour, rug, candlestick, ohlc, scatter_matrix, 
bubble, heatmap, imshow, sunburst, treemap, histogram2d, density_heatmap, choropleth, scattergeo, scattermapbox, 
density_mapbox, waterfall, funnel, sankey, timeline, indicator, scatter_3d, surface, line_3d, mesh3d


Output JSON format:
{
  "analysis_options": [
    {
      "name": "Trend Analysis",
	  "chart_type": "line",
      "description": "Detects overall trends in time-series data.",
      "columns": ["<Relevant Time Column>"]
    },
    {
      "name": "Correlation Matrix",
	  "chart_type": "heatmap",
      "description": "Shows correlations between numerical columns.",
      "columns": ["<Numerical Column 1>", "<Numerical Column 2>"]
    },
    ...
  ]
}

Example Output:
{
  "analysis_options": [
    {
      "name": "Distribution Analysis",
	  "chart_type": "bar",
      "description": "Shows the distribution of values in a categorical column.",
      "columns": ["Category"]
    },
    {
      "name": "Seasonality Analysis",
	  "chart_type": "line",
      "description": "Identifies seasonal patterns in time-series data.",
      "columns": ["Date"]
    },
    ...
  ]
}
`,
			},
			{
				Role: "user",
				Content: fmt.Sprintf(`The shape of the data:
%s
%s

Analysis Instructions:
- Review the data structure, column types, and any relationships between columns.
- Select 8 impactful analyses from the analysis catalog, focusing on the most suitable options for the given data type (e.g., time series, categorical, numerical, etc.).
- For each selected analysis, identify the ideal chart type (e.g., line chart for time series trends, bar chart for categorical frequency).

Respond with a JSON object in the following format, where you insert the Options based on the data:
{
  "analysis_options": [
    {
      "name": "Trend Analysis",
	  "chart_type": "line",
      "description": "Detects overall trends in time-series data.",
      "columns": ["<Relevant Time Column>"]
    },
    {
      "name": "Correlation Matrix",
	  "chart_type": "heatmap",
      "description": "Shows correlations between numerical columns.",
      "columns": ["<Numerical Column 1>", "<Numerical Column 2>"]
    },
    ...
  ]
}`,
					data.HeadersString(),
					data.FirstRowsString(),
				),
			},
		},
	}
}
