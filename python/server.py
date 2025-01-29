import asyncio
from fastapi import FastAPI, Request, Query
from fastapi.responses import StreamingResponse
from transformers import AutoModelForCausalLM, AutoTokenizer, TextIteratorStreamer
from threading import Thread
import logging
import torch

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

logger.info("Loading GPT-2 model and tokenizer...")
model = AutoModelForCausalLM.from_pretrained("gpt2")
tokenizer = AutoTokenizer.from_pretrained("gpt2")

app = FastAPI()


@app.post("/start")
async def start_stream(request: Request):
    raw_body = await request.body()
    print(f"Raw request body received: {raw_body}")

    try:
        data = await request.json()
        message = data.get("message", "").strip()
        if not message:
            return {"error": "Message cannot be empty"}
        return {"status": "Message received", "message": message}
    except Exception as e:
        return {"error": f"Invalid request format: {e}"}


@app.get("/generate")
async def generate(message: str = Query(..., description="User input message")):
    """Generates text based on the input message."""
    logger.info(f"Received request at /generate with message: {message}")

    try:
        inputs = tokenizer([message], return_tensors="pt")
        streamer = TextIteratorStreamer(tokenizer)

        generation_kwargs = dict(inputs, streamer=streamer, max_new_tokens=20)

        logger.info("Starting text generation thread...")
        thread = Thread(target=model.generate, kwargs=generation_kwargs)
        thread.start()

        async def stream():
            logger.info("Streaming response to client...")
            try:
                for new_text in streamer:
                    logger.info(f"Generated token: {new_text.strip()}")
                    await asyncio.sleep(0.2)
                    yield new_text + "\n"
                logger.info("Text generation complete. Streaming finished.")
            except Exception as e:
                logger.error(f"Error during streaming: {e}")

        return StreamingResponse(stream(), media_type="text/plain")

    except Exception as e:
        logger.error(f"Error processing /generate request: {e}")
        return {"error": "An error occurred during text generation"}


if __name__ == "__main__":
    import uvicorn

    logger.info("Starting FastAPI server...")
    uvicorn.run(app, host="0.0.0.0", port=8000)
