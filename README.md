# Streaming Text Generation Application

This project is a text generation application that streams generated text from a Python FastAPI server to a Go-based backend, which then serves it to a web frontend using Server-Sent Events (SSE).

## Overview

- **Python FastAPI Server (`server.py`)**: Uses Hugging Face's `transformers` library to generate text with GPT-2 and streams it as responses.
- **Go Backend (`main.go`)**: Acts as a bridge, requesting text generation from FastAPI and streaming results to the frontend using SSE.
- **HTML & JavaScript Frontend (`home.html`)**: Displays streamed text in real-time.

## How It Works

1. The user clicks the **Start** button on the web interface.
2. A request is made to the `/start` endpoint of the Go server.
3. The Go server calls the FastAPI server (`http://127.0.0.1:8000`) to initiate text generation.
4. FastAPI streams the generated text token by token.
5. The Go server reads these tokens and forwards them as SSE to the frontend.
6. The frontend listens for updates and displays the generated text dynamically.

## Setup Instructions

### Prerequisites
- Python 3.8+
- Go 1.16+
- `pip install fastapi transformers uvicorn`

### Running the Application

1. **Start the FastAPI server** (Python):
   ```sh
   python server.py
   ```

2. **Start the Go server**:
   ```sh
   go run main.go
   ```

3. **Open the application** in your browser at:
   ```
   http://localhost:8000
   ```

## File Structure

```
project/
│── python/
│   ├── server.py            # FastAPI server for text generation
│── go/
│   ├── main.go              # Go backend with SSE implementation
│   ├── templates/
│       ├── home.html        # Frontend page
```

## API Endpoints

### FastAPI Server

- `GET /` → Streams generated text.

### Go Server

- `POST /start` → Starts text generation.
- `GET /generate` → Streams generated text to the frontend via SSE.

## Technologies Used

- **Python (FastAPI, Transformers, Uvicorn)**
- **Go (Gorilla Mux, HTTP Server)**
- **JavaScript (EventSource for SSE)**

## Notes

- Ensure both servers are running before accessing the web UI.
- Modify `model` in `server.py` to change text generation behavior.

