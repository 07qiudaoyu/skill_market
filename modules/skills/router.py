from fastapi import APIRouter, Depends, status

from .schemas import SkillSearchResponse
from .service import SkillsService
from .dependencies import get_skill_service

router = APIRouter(
    prefix="/skills",
    tags={"Skills"}
)

@router.get(
    "/search",
    response_model=SkillSearchResponse,
    status_code=status.HTTP_200_OK
)
def search_skills(
    q: str | None = None,
    category: str | None = None,
    tags: str | None = None,
    sort: str = "newest",
    page: int = 1,
    size: int = 10,
    service: SkillsService = Depends(get_skill_service)
):
    return service.search_skills(
        q=q,
        category=category,
        tags=tags,
        sort=sort,
        page=page,
        size=size
    )