# Use an official Python runtime as a parent image
FROM python:3.9-slim

# Set the working directory
WORKDIR /app

# Copy the current directory contents into the container
COPY . .

# Install basic system dependencies commonly used for small data analysis
RUN apt-get update && apt-get install -y \
    libxml2-dev \
    libxslt-dev \
    libz-dev \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install essential Python libraries for data manipulation and Plotly visualization with pinned versions
RUN pip install --no-cache-dir \
    fastapi==0.95.2 \
    uvicorn==0.22.0 \
    numpy==1.23.5 \
    pandas==1.5.3 \
    plotly==5.11.0 \
    scikit-learn==1.2.2

# Install additional data handling libraries for small datasets with pinned versions
RUN pip install --no-cache-dir \
    openpyxl==3.1.2 \
    xlrd==2.0.1 \
    lxml==4.9.2 \
    beautifulsoup4==4.11.1

# Install lightweight data visualization libraries if needed (optional)
RUN pip install --no-cache-dir \
    bokeh==3.0.3 \
    folium==0.14.0 \
    geopandas==0.13.2 \
    missingno==0.5.2

# Clean up unnecessary packages to reduce image size
RUN apt-get purge -y libxml2-dev libxslt-dev && apt-get autoremove -y

# Expose port 8000 for the API
EXPOSE 8000

# Command to run the FastAPI app with Uvicorn
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]

