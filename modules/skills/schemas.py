
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