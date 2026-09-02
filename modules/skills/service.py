from .repository import SkillsRepository
from fastapi import HTTPException
import os
from fastapi.responses import FileResponse
from fastapi.responses import StreamingResponse
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
    def Se_Download_Zip(self,skill_versions_id:int):#本地的
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
        return StreamingResponse(
            file_download(file_path),
            media_type="application/zip",
            headers={
               "Content-Disposition": f'attachment; filename="{os.path.basename(file_path)}"'
            }
        )

               