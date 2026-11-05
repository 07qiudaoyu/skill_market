
from datetime import datetime
from pydantic import BaseModel
from enum import Enum
from fastapi import UploadFile
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
    id:int
    skill_name:str
    version:str
    size_bytes:int
    extract_file_bytes:int | None = None#解压后大小，老数据可能为空
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
    #----------------------------------------------
class Sc_Download_Zip_Request(BaseModel):
    id:int
   # token:str

#--------------------------------------------
class Sc_Skill_Tags(str,Enum):
    python="python"
    javascript="javascript"
    email="email"
    crawler="crawler"
    api="api"
    ai="ai"
    webhook="webhook"
    report="report"
    excel="excel"
    pdf="pdf"
class Sc_Skill_Categories(str,Enum):
    web_dev="web-dev"
    data_ai="data-ai"
    automation="automation"
    productivity="productivity"
    devops="devops"
#----------------------------------------------
class Sc_Upload_Zip_request(BaseModel):
    name:str
    version:str
   # token:str
    category:Sc_Skill_Categories
    tags:Sc_Skill_Tags
    readme_html:str
    summary:str
    slug:str
class Sc_Upload_Zip_response(BaseModel):
    message: str
    file_name: str
    storage_key: str

  

