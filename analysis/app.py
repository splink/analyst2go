from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
import base64
import io
import traceback

app = FastAPI()

# Define the request model for receiving Python code
class CodeRequest(BaseModel):
    code: str

@app.post("/generate-chart/")
async def generate_chart(request: CodeRequest):
    # Create a dictionary to safely execute user code
    exec_globals = {"plt": plt, "pd": pd, "np": np, "output": None}
    code = request.code

    try:
        # Execute the user-provided code in a controlled namespace
        exec(code, exec_globals)

        # Check if 'output' variable exists and is a matplotlib figure
        if isinstance(exec_globals["output"], plt.Figure):
            fig = exec_globals["output"]
        else:
            raise ValueError("No valid matplotlib figure found in 'output'.")

        # Convert figure to SVG
        buf = io.BytesIO()
        fig.savefig(buf, format="svg")
        buf.seek(0)

        # Encode SVG to base64
        img_base64 = base64.b64encode(buf.getvalue()).decode("utf-8")
        return {"chart": img_base64, "format": "svg"}

    except Exception as e:
        error_message = f"Error in executing code: {str(e)}\n{traceback.format_exc()}"
        raise HTTPException(status_code=400, detail=error_message)
    finally:
        plt.close("all")  # Close all figures to free up memory

