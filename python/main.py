import os
import logging
import pathlib
import hashlib
from fastapi import FastAPI, Form, HTTPException, File, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from typing import List

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "images"
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

# 商品を登録する空のリストを作成
items = []

image_dir = "images"
os.makedirs(image_dir,exist_ok = True)

@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    image_cont = image.file.read()
    sha256_hash = hashlib.sha256(image_cont).hexdigest()
    image_file = f"{sha256_hash}.jpg"
    image_path = os.path.join(image_dir, image_file)

    with open(image_path, "wb") as image_file:
        image_file.write(image_cont)

    item = {"name": name, "category": category, "image_name": image_file}
    items.append(item)
    logger.info(f"Receive item: {name}")

    with open("items.json", "w") as f:
        import json
        json.dump(items,f)
    return {"message": f"item received: {name}", "image_name": image_file}

@app.get("/items")
def get_items():
    return {"items": items}

@app.get("/image/{image_name}")
async def get_image(image_name: str):
    # Create image path
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)
