from fastapi import FastAPI
from api.router import api_router  # 使用完整路径



app = FastAPI(
    title="Skill Market API"
)


app.include_router(api_router)


