from fastapi import APIRouter, Depends, status

from .schemas import SkillSearchResponse,Find_ZIP_Response
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
@router.get(
        "/find_versions",
        response_model=Find_ZIP_Response,
        status_code=status.HTTP_200_OK
)
def FIND_SKILL_VERSIONS(
        public_id:str,
        service: SkillsService=Depends(get_skill_service)
):
    return service.find_skill_versions(
        public_id=public_id
    )
@router.get(
    "/request_file",
    status_code=status.HTTP_200_OK
)
def REQUEST_FILE(
    skill_versions_id:int,
    service:SkillsService=Depends(get_skill_service)
):
    return service.Se_Find_Zip(
        skill_versions_id=skill_versions_id
    )
@router.get(
    
    "/Download_file",
    status_code=status.HTTP_200_OK
)
def DOWNLOAD_ZIP(
    id:int,
    service:SkillsService=Depends(get_skill_service)
):
    return service.Se_Download_Zip(
        skill_versions_id=id
    )