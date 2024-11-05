import traceback
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import plotly.io as pio
import pandas as pd
import numpy as np

app = FastAPI()

# Define the request model for receiving Python code
class CodeRequest(BaseModel):
    code: str

@app.post("/generate-chart/")
async def generate_chart(request: CodeRequest):
    # Create a dictionary to safely execute user code
    exec_globals = {"pd": pd, "np": np, "px": pio, "output": None}
    code = request.code

    try:
        # Execute the user-provided code in a controlled namespace
        exec(code, exec_globals)

        # Check if 'output' variable exists and is a Plotly figure
        if "output" in exec_globals and exec_globals["output"] is not None:
            fig = exec_globals["output"]
        else:
            raise ValueError("No valid Plotly figure found in 'output'.")

        # Convert Plotly figure to JSON
        fig_json = fig.to_json()
        return {"chart": fig_json, "format": "plotly_json"}

    except Exception as e:
        error_message = f"Error in executing code: {str(e)}\n{traceback.format_exc()}"
        raise HTTPException(status_code=400, detail=error_message)

