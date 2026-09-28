from pathlib import Path

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from api.router import api_router  # 使用完整路径


app = FastAPI(
    title="Skill Market API"
)


app.include_router(api_router)

# 前端静态页：挂在根路径，与 /api/v1 同源，页面里的 fetch 即可直接调接口
web_dir = Path(__file__).resolve().parent.parent / "web"
if web_dir.is_dir():
    app.mount("/", StaticFiles(directory=web_dir, html=True), name="web")


