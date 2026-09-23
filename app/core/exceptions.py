#异常代码的文件处理
from fastapi import HTTPException
#用户上传skills，slug已被使用处理
class Ex_upload_file:
   def Ex_Upload_slug_repeat():
       raise HTTPException(
                status_code=404,
                detail="该slug已被建立,请重新换一个"
            )
   def Ex_Upload_size_fileNO():
       raise HTTPException(
           status_code=413,
           detail="上传zip文件过大"
       )
   def Ex_lack_json():
       raise HTTPException(
                  status_code=400,
                  detail="上传zip文件，缺少skill.json"
              )
   def Ex_lack_md():
       raise HTTPException(
                         status_code=400,
                         detail="上传zip文件，缺少SKILL.md"
                     )
   def Ex_lack_read():
       raise HTTPException(
                         status_code=400,
                         detail="上传zip文件，缺少README.md"
                     )
   def Ex_really_zip():
       raise HTTPException(
                                status_code=413,
                                detail="上传zip文件，zip解压内存过大"
                            )
   def Ex_really_json():
       raise HTTPException(
           status_code=400,
         detail="上传zip文件，skill.json数据与上传审核不通过"
       )
   def Ex_invalid_encoding():
       raise HTTPException(
                  status_code=400,
                detail="上传zip文件，SKILL.md文件要可转化为utf-8"
              )
   def Ex_invalid_skill():
       raise HTTPException(
                         status_code=400,
                       detail="上传zip文件，SKILL.md文件有符合数据与上传一致"
                     )
   def Ex_really_skill():
       raise HTTPException(
                                       status_code=413,
                                       detail="上传zip文件,description或name过长"
                                   )
   def Ex_user_ids():
       raise HTTPException(
           status_code=400,
           detail="该slug已被其他用户使用"
       )
   def Ex_insert_zip():
       raise HTTPException(
                  status_code=404,
                  detail="数据库添加失败"
              )
       