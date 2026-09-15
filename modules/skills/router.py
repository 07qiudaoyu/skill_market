from fastapi import APIRouter, Depends, status

from .schemas import SkillSearchResponse,Find_ZIP_Response,Sc_Upload_Zip_response,Sc_Skill_Categories
from .service import SkillsService
from .dependencies import get_skill_service
from fastapi import File, Form, UploadFile

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
    token:str,
    service:SkillsService=Depends(get_skill_service)
):
    return service.Se_Download_Zip(
        skill_versions_id=id,
        token=token
    )

@router.post(
    "/upload_zip",
    response_model=Sc_Upload_Zip_response,
    status_code=status.HTTP_200_OK
)
async def Ro_Upload_Zip(
    name: str = Form(...),
    version: str = Form(...),
    token: str = Form(...),
    category: Sc_Skill_Categories = Form(...),
    tags: str = Form(...),
    readme_html: str = Form(...),
    summary: str = Form(...),
    slug: str = Form(...),
    
    upload_zip: UploadFile = File(...),
    service:SkillsService=Depends(get_skill_service)
):
    return await service.SE_Upload_Zip(
        name=name,
        version=version,
        token=token,
        category=category.value,
        tags=tags,
        readme_html=readme_html,
        summary=summary,
        slug=slug,
        upload_zip=upload_zip
    )