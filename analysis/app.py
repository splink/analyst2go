from fastapi import FastAPI, HTTPException, File, Form, UploadFile
from pydantic import BaseModel
import pandas as pd
import numpy as np
import plotly.express as px
import plotly.graph_objects as go
import plotly.io as pio
import traceback
import logging
import io

app = FastAPI()

class CodeRequest(BaseModel):
    code: str

@app.post("/generate-chart/")
async def generate_chart(code: str = Form(...), file: UploadFile = File(...)):
    # Read file into a Pandas DataFrame
    file_content = await file.read()
    file_extension = file.filename.split(".")[-1]
    if file_extension == "csv":
        df = pd.read_csv(io.BytesIO(file_content))
    elif file_extension in ["xls", "xlsx"]:
        df = pd.read_excel(io.BytesIO(file_content))
    else:
        raise ValueError("Unsupported file format. Please upload a CSV or Excel file.")

    # Clean data
    df.dropna(how="all", inplace=True)
    df.dropna(axis=1, how="all", inplace=True)

    df.columns = df.columns.str.strip().str.lower().str.replace(' ', '_')

    for col in df.columns:
        if 'date' in col or 'time' in col:
            df[col] = pd.to_datetime(df[col], errors='coerce')

    for col in df.select_dtypes(include=["number"]).columns:
        df[col].fillna(df[col].mean(), inplace=True)
    for col in df.select_dtypes(include=["object"]).columns:
        df[col].fillna(df[col].mode()[0], inplace=True)

    for col in df.select_dtypes(include=["int", "float"]).columns:
        df[col] = pd.to_numeric(df[col], errors='coerce')
    for col in df.select_dtypes(include=["object"]).columns:
        df[col] = df[col].astype(str)

    df.drop_duplicates(inplace=True)
    df.reset_index(drop=True, inplace=True)

    # Define the execution environment
    exec_globals = {"pd": pd, "np": np, "px": px, "go": go, "df": df, "output": None}

    try:
        # Execute the code with the file data in a controlled namespace
        exec(code, exec_globals)

        # Retrieve the Plotly figure from the execution environment
        if "output" in exec_globals and isinstance(exec_globals["output"], go.Figure):
            fig = exec_globals["output"]
        else:
            raise ValueError("No valid Plotly figure found in 'output'.")

        # Convert the Plotly figure to JSON and return it
        fig_json = fig.to_json()
        return {"chart": fig_json, "format": "plotly_json"}

    except Exception as e:
        error_message = f"Error in executing code: {str(e)}\n{traceback.format_exc()}"
        logging.error(error_message)
        raise HTTPException(status_code=400, detail=error_message)
