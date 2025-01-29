import asyncio
from fastapi import FastAPI
from fastapi.responses import StreamingResponse
from transformers import AutoModelForCausalLM, AutoTokenizer, TextIteratorStreamer
from threading import Thread

model = AutoModelForCausalLM.from_pretrained("gpt2")
tokenizer = AutoTokenizer.from_pretrained("gpt2")

app = FastAPI()


async def generate():
    inputs = tokenizer(["An increasing sequence: one,"], return_tensors="pt")
    streamer = TextIteratorStreamer(tokenizer)

    generation_kwargs = {
        **inputs,
        "streamer": streamer,
        "max_new_tokens": 50,
        "pad_token_id": tokenizer.eos_token_id,
    }

    thread = Thread(target=model.generate, kwargs=generation_kwargs)
    thread.start()

    for new_text in streamer:
        await asyncio.sleep(0.2)
        yield new_text + "\n"


@app.get("/")
async def main():
    return StreamingResponse(generate(), media_type="text/plain")


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)
