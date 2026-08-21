'''你可以把它理解成：

总路由表。'''
from fastapi import APIRouter

from modules.auth.router import router as auth_router


api_router = APIRouter(prefix="/api/v1")

api_router.include_router(auth_router)