from zipfile import ZipFile
from fastapi import UploadFile
from io import BytesIO
import os
from app.core.exceptions import Ex_upload_file
async def Check_file_zip(upload_zip:UploadFile):
    file_content= await upload_zip.read()
    size_bytes= len(file_content)
    #1.检查zip大小
    if size_bytes>52428800:
        Ex_upload_file.Ex_Upload_size_fileNO()#50MB
    await upload_zip.seek(0)
    #2.检查文件是否有skill.md,skill.json,readme.md
    zip_file=ZipFile(BytesIO(file_content))#解压
    file_names=[os.path.basename(n) for n in zip_file.namelist()]
    print("DEBUG file_names =", file_names) 
    if "skill.json" not in file_names:
        Ex_upload_file.Ex_lack_json()
    if "SKILL.md" not in file_names:
        Ex_upload_file.Ex_lack_md()
    # 缺少 SKILL.md
    if "README.md" not in file_names:
        Ex_upload_file.Ex_lack_read()
    #3.检查解压的内存是否过大
    uncompressed_size = sum(
    info.file_size
    for info in zip_file.infolist()
    )
    if uncompressed_size>(100*1024*1024):#100MB
        Ex_upload_file.Ex_really_zip()
    return {
    "zip_size": size_bytes,
    "uncompressed_size": uncompressed_size,
       "file_content": file_content,
}


