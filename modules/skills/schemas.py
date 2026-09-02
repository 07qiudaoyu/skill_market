
from datetime import datetime
from pydantic import BaseModel


class SkillSearchParams(BaseModel):
    q: str | None       
    category: str | None 
    tags: list[str] | None 
    sort: str | None    
    page: int            
    size: int      
class SkillListItem(BaseModel):#在我的搜索接口使用下，这个是起文本的规范，字面上的作用，tags相对get接口而言str对用户是最合适
    public_id: str
    slug: str
    display_name: str
    summary: str
    category_name: str | None
    tags: list[str] | None
    download_count: int
    rating_avg: float
    created_at: datetime
class SkillSearchResponse(BaseModel):
    message: str
    items: list[SkillListItem]
    page: int
    size: int
class All_Skill_Versions(BaseModel):
    skill_name:str
    version:str
    size_bytes:int
    file_name:str
    created_at:datetime
    founder_name:str
class Find_ZIP_Request(BaseModel):
    public_id:str
class Find_ZIP_Response(BaseModel):
    message:str
    versions:list[All_Skill_Versions]
class Sc_Find_Zip_Request(BaseModel):
    skill_versions_id:int
class Sc_Download_Zip_Request(BaseModel):
    id:int

  

