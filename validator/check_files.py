from zipfile import ZipFile
from fastapi import UploadFile
from io import BytesIO
import json
import yaml

from app.core.exceptions import Ex_upload_file
async def Check_Files(file_content:bytes,
                      tags:list[str],
                      category:str,
                        version: str,
                        summary: str,
                        slug:str):
    zip_file=ZipFile(BytesIO(file_content))
    skill_json=zip_file.read("skill.json")
    skill_md=zip_file.read("SKILL.md")
    
    #skill_json元数据进行核对
    skill_json_date=json.loads(skill_json)
    if slug!=skill_json_date["name"]\
    or sorted(tags)!=sorted(skill_json_date["tags"]) \
    or version!=skill_json_date["version"]\
    or category!=skill_json_date["category"]\
    or summary!=skill_json_date["description"]:
        Ex_upload_file.Ex_really_json()
    try:
        skill_md_data=skill_md.decode("utf-8")
    except UnicodeDecodeError:
        return Ex_upload_file.Ex_invalid_encoding()#skill.md文件是否能传为utf-8
    
    parts = skill_md_data.split("---", 2)
    if len(parts)< 3 or parts[0].strip() != "":
     return Ex_upload_file.Ex_invalid_skill()
    skill_md_data1 = yaml.safe_load(parts[1])
    if slug != skill_md_data1["name"]\
    or summary != skill_md_data1["description"]:
        return Ex_upload_file.Ex_invalid_skill()#是否有name等且是否相同
    if not 1<=len(skill_md_data1["description"])<=1024:
        return Ex_upload_file.Ex_really_skill()
    if not 1 <= len(skill_md_data1["name"]) <= 64:
       return Ex_upload_file.Ex_really_skill()
    return {
    "success": True,
    "file_content": file_content,
}
        

    
