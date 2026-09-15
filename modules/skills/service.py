from .repository import SkillsRepository
from fastapi import HTTPException
import os
from fastapi.responses import FileResponse
from fastapi.responses import StreamingResponse
from app.core.security import check_token
from pathlib import Path
from fastapi import UploadFile
from app.core.security import check_token
class SkillsService:
    def __init__(self, repository: SkillsRepository):
        self.repository = repository

    def search_skills(self, q=None, category=None, tags=None, sort="newest", page=1, size=10):
        items = self.repository.select_skills_re(
            q=q, category=category, tags=tags,
            sort=sort, page=page, size=size
        )
        return {
            "message": "搜索成功",
            "items": items,
            "page": page,
            "size": size
        }
    def find_skill_versions(self,public_id:str):
        result = self.repository.Find_skill_versions(
            public_id=public_id
        )
        if not result:
          raise HTTPException(
            status_code=404,
            detail="该技能不存在或未发布版本"
        )
        return {
"message":"该技能所有版本及发表用户",
"versions":result
        }
    def Se_Find_Zip(self,skill_versions_id:int):
        result=self.repository.Re_Find_Zip(
            skill_versions_id=skill_versions_id
        )
        if not result:
            raise HTTPException(
                status_code=404,
                detail="未找到skill_versions_id对应信息,错误"
            )
        STORAGE_ROOT="data"
        file_path=os.path.join(
            STORAGE_ROOT,
            result["storage_key"]
        )
        if not os.path.isfile(file_path):
            raise HTTPException(
                status_code=404,
                detail="未找到skill_versions_id对应的文件地址,错误"
            )
        return FileResponse(
            file_path,
            filename=os.path.basename(file_path)
        )
    def Se_Download_Zip(self,skill_versions_id:int,token:str):#本地的
        payload=check_token(token)
        
        if payload is None:
                    raise HTTPException(
                        status_code=401,
                        detail="token错误!游客现不支持下载skill,请登入"
                    )
        user_id=payload["id"]
        result=self.repository.Re_Find_Zip(skill_versions_id=skill_versions_id)
        if not result:
            raise HTTPException(status_code=404,detail="未找到版本信息")
        file_path=os.path.join("data",result["storage_key"])
        if not os.path.isfile(file_path):
            raise HTTPException(status_code=404,detail="未找到对应的zip文件")
        def file_download(path,chunk_size=8192):
            with open(path,"rb") as f:
                while True:
                    chunk=f.read(chunk_size)
                    if not chunk:
                        break
                    yield chunk
        self.repository.Re_Download_zip(user_id,skill_versions_id)
        return StreamingResponse(
            file_download(file_path),
            media_type="application/zip",
            headers={
               "Content-Disposition": f'attachment; filename="{os.path.basename(file_path)}"'
            }
        )
    async def SE_Upload_Zip(self,
                    upload_zip: UploadFile,
                    name: str,
                    version: str,
                    token:str,
                    category: str,
                    tags: str,
                    readme_html: str,
                    summary: str,
                    slug: str
        ):
        #解码
        result1=check_token(token)
        if result1 is None:
                    raise HTTPException(
                        status_code=401,
                        detail="token错误!"
                    )
        user_id=result1["id"]
        #将str变为list
        change_tags=tags.split(",")
        #读取上传文件内容，计算大小
        file_content=await upload_zip.read()
        size_bytes=len(file_content)
        result=self.repository.Re_Upload_Zip(name=name,
                                             version=version,
                                             user_id=user_id,
                                             category=category,
                                             readme_html=readme_html,
                                             summary=summary,
                                             slug=slug,
                                             change_tags=change_tags,
                                             size_bytes=size_bytes
                                             )
        #这里就开始写上传的函数操作
        file_path= Path("data") / result["storage_key"]
        file_name=result["file_name"]
        headers={
            "Content-Disponsition":f'attachment; filename="{file_name}"'
        }
        file_path.parent.mkdir(
    parents=True,
    exist_ok=True
)
        with file_path.open("wb") as buffer:
         buffer.write(file_content)
        return {
        "message": "文件上传成功",
        "file_name": file_name,
        "storage_key": result["storage_key"]
    }
        
               
    
               