from fastapi import APIRouter
from modules.auth.router import router as auth_router
from modules.skills.router import router as skills_router


api_router = APIRouter(prefix="/api/v1")
api_router.include_router(auth_router)
api_router.include_router(skills_router)

