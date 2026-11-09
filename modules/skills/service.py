from .repository import SkillsRepository
from fastapi import HTTPException
import os
from fastapi.responses import StreamingResponse
from app.core.security import check_token
from pathlib import Path
from fastapi import UploadFile
from app.core.security import check_token
import requests
import json
from GitHub_validator import check_GitHub_file
from validator import check_file_zip
import re
import markdown
import zipfile
import io
from starlette.datastructures import UploadFile as StarletteUploadFile


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
    #----------------------------------------------------
    def Se_Download_Zip(self,skill_versions_id:int,payload:dict):#本地的
        
        
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
                    #文件写完了，我们就直接进行更新
            try:
                 ak=self.repository.Re_Uerts_download(user_id,skill_versions_id)#
                 if ak==1:
                      print("第一次下载")
                 else:
                      print("非第一次下载")
            except Exception as e:
                 print("下载成功，数据库更新失败:",e)
                     
        
        return StreamingResponse(
            file_download(file_path),
            media_type="application/zip",
            headers={
               "Content-Disposition": f'attachment; filename="{os.path.basename(file_path)}"'
            }
        )
    #==========================================================
    async def SE_Upload_Zip(self,
                    upload_zip: UploadFile,
                    name: str,
                    version: str,
                    payload:dict,
                    category: str,
                    tags: list[str],
                    readme_html: str,
                    summary: str,
                    slug: str
        ):
        #解码
        result1=payload
        if result1 is None:
                    raise HTTPException(
                        status_code=401,
                        detail="token错误!"
                    )
        user_id=result1["id"]
        #将str变为list
    
        #读取上传文件内容，计算大小
        
        result=await self.repository.Re_Upload_Zip(upload_zip=upload_zip,
                                             name=name,
                                             version=version,
                                             owner_id=user_id,
                                             category=category,
                                             readme_html=readme_html,
                                             summary=summary,
                                             slug=slug,
                                             tags=tags
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
         buffer.write(result["file_content"])
        return {
        "message": "文件上传成功",
        "file_name": file_name,
        "storage_key": result["storage_key"]
    }
    #===========================================






         
    def Se_All_tags(self):#按步骤来最好，现在返回所有的tags
         return self.repository.Re_All_Tags()
    def Se_All_categories(self):
         return self.repository.Re_All_categories()
    
    #=========================================GitHub的sdk
    async def SE_GitHub_post(self,url1:str,payload:dict
                       ):
         #1.先审核jwt:
         if payload is None:
            raise HTTPException(
            status_code=401,
           detail="token错误!游客现不支持下载skill,请登入"
                   )
         
         #白名单https://github.com/{owner}/{repo}
         def url_extract(url1):
              result=r"https?://github\.com/([^/]+)/([^/]+?)(?:\.git)?/?"
              result1 = re.fullmatch(result, url1)
              if not result1:
                   raise HTTPException(
                        status_code=401,
                        detail="用户上传url不符合要求https://github.com/{owner}/{repo}"
                   )
              owner=result1.group(1)
              repo=result1.group(2).removesuffix(".git")#取除末尾的后缀
              main_url=f"https://api.github.com/repos/{owner}/{repo}/zipball"
              return {"main_url":main_url,
                      "owner":owner,
                      "repo":repo}
         url_datas=url_extract(url1)
         main_url=url_datas["main_url"]
         owner=url_datas["owner"]
         repo=url_datas["repo"]

         #2.带有文件要求和文件数量要
         GITHUB_TOKEN = os.environ.get("GITHUB_TOKEN")
         headers = {"Authorization": f"Bearer {GITHUB_TOKEN}"}
         r=requests.get(main_url,headers=headers,timeout=30)
         check_GitHub_file.Ex_GitHUb_main.Ex_GitHub_status_code(r.status_code)#请求

         #-------------- 去目录               --------------------
         raw = zipfile.ZipFile(io.BytesIO(r.content))#将zip打开，放在临时内存
         prefix = raw.namelist()[0].split("/")[0] + "/"
         buf = io.BytesIO()
         with zipfile.ZipFile(buf, "w", zipfile.ZIP_DEFLATED) as out:
              for item in raw.infolist():
                   entry = item.filename[len(prefix):]
                   if not entry or entry.endswith("/"):
                        continue
                   out.writestr(entry, raw.read(item))
         zip_bytes = buf.getvalue()
         #审核：根目录三件套必须在
         check_up=StarletteUploadFile(filename="check.zip",file=io.BytesIO(zip_bytes))
         await check_file_zip.Check_file_zip(check_up)
         #从skill.json拿元数据
         src=zipfile.ZipFile(io.BytesIO(zip_bytes))
         data=json.loads(src.read("skill.json"))
         result1=payload
         if result1 is None:
                             raise HTTPException(
                                 status_code=401,
                                 detail="token错误!"
                             )
         user_id=result1["id"]
         slug=data["name"]
         version =data["version"]
         category= data["category"]
         tags= data["tags"]
         summary=data["description"]
         readme_html=markdown.markdown(src.read("README.md").decode("utf-8"))

         upload=StarletteUploadFile(
              filename=f"{slug}.zip",
              file=io.BytesIO(zip_bytes)
         )
         name=data.get("displayName") or slug
         result=await self.repository.Re_Upload_Zip(upload_zip=upload,
                                                      name=name,
                                                      version=version,
                                                      owner_id=user_id,
                                                      category=category,
                                                      readme_html=readme_html,
                                                      summary=summary,
                                                      slug=slug,
                                                      tags=tags
                                                      )
         #3.把zip落到磁盘，下载接口才有东西可读
         file_path=Path("data") / result["storage_key"]
         file_path.parent.mkdir(parents=True, exist_ok=True)
         file_path.write_bytes(result["file_content"])





                  
         

         





         ak=r.headers.get("x-ratelimit-remaining")
         return {
              "message":"ok",
              "limit":ak 
         }
         