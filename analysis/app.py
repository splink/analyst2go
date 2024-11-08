from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import pandas as pd
import numpy as np
import plotly.express as px
import plotly.graph_objects as go
import plotly.io as pio
import traceback
import logging


app = FastAPI()

class CodeRequest(BaseModel):
    code: str

@app.post("/generate-chart/")
async def generate_chart(request: CodeRequest):
    # Log the received code for debugging
    logging.debug(f"Received code:\n{request.code}")

    # Define a controlled namespace for executing user code
    exec_globals = {"pd": pd, "np": np, "px": px, "go": go, "pio": pio, "output": None}
    code = request.code

    try:
        # Execute the user-provided code in a restricted environment
        exec(code, exec_globals)

        # Check if 'output' variable is defined and is a Plotly figure
        if "output" in exec_globals and isinstance(exec_globals["output"], go.Figure):
            fig = exec_globals["output"]
        else:
            raise ValueError("No valid Plotly figure found in 'output'.")

        # Convert the Plotly figure to JSON
        fig_json = fig.to_json()
        return {"chart": fig_json, "format": "plotly_json"}

    except Exception as e:
        # Log the full traceback for debugging
        error_message = f"Error in executing code: {str(e)}\n{traceback.format_exc()}"
        logging.error(error_message)
        raise HTTPException(status_code=400, detail=error_message)
