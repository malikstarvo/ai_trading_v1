import asyncio
from pathlib import Path
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from fastapi.middleware.cors import CORSMiddleware

from api.db import db
from api.routers import (
    health,
    market,
    features,
    trading,
    model,
    system,
    ws,
    data,
    actions,
)


@asynccontextmanager
async def lifespan(app: FastAPI):
    await db.connect()
    task = asyncio.create_task(ws.broadcast_loop(10.0))
    yield
    task.cancel()
    await db.close()


app = FastAPI(lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(health.router)
app.include_router(market.router)
app.include_router(features.router)
app.include_router(trading.router)
app.include_router(model.router)
app.include_router(system.router)
app.include_router(data.router)
app.include_router(actions.router)
app.include_router(ws.router)  # WebSocket before static mount


@app.get("/api/healthz")
async def healthz():
    return {"status": "ok"}


static_dir = Path(__file__).resolve().parent.parent / "web" / "dist"
if static_dir.is_dir():
    app.mount("/", StaticFiles(directory=str(static_dir), html=True), name="static")
