# Use Python 3.11 slim image for smaller size
FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Copy requirements file first for better caching
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy the application code
COPY main.py .

# Set environment variables with sensible defaults
ENV REDIS_HOST=redis \
    REDIS_PORT=6379 \
    PUBLISH_INTERVAL=2.0

# Define the entrypoint
CMD ["python", "main.py"]