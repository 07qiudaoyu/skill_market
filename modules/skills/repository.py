
from app.core.exceptions import Ex_upload_file
from validator import check_file_zip,check_files
from fastapi import UploadFile
from zipfile import ZipFile
from fastapi import HTTPException
MAX_file_zip=52428800#也就是50MB的内存大小，这个是为解压的文件
class SkillsRepository:
    def __init__(self, conn):
        self.conn = conn

    def select_skills_re(self, q: str | None = None,
                     category: str | None = None,
                     tags: str | None = None,
                     sort: str = "newest",
                     page: int = 1,
                     size: int = 10):
        cursor = self.conn.cursor()
        condition=[]
        params=[]
        tags=tags.split(",") if tags else []
        if q:    
                    condition.append("(s.display_name like %s or s.summary like %s)")
                    params.extend([f"%{q}%",f"%{q}%"])#一次行放多个
        if category:
                    condition.append("c.slug =%s")
                    params.append(category)#一次行放一个
        if tags:
                    ak=",".join(["%s"]*len(tags))
                    #我的tags时str形式的输入，不是list，不要与schemas文本规范作用混淆了
                    condition.append(f"""
        s.id in(select skill_id from skill_tags st
        join tags t on st.tag_id=t.id
        where t.slug in ({ak})
        group by skill_id
        having count(distinct t.slug)=%s
        )
        """)#ak时先确保slug找到，而后的having是确保数量，也就是都找到
                    params.extend(tags)
                    params.append(len(tags))
        where_sql=""
        if condition:
                        where_sql=" where " + " and ".join(condition)
        sort_map={
                        "newest":"s.created_at desc",
                        "downloads":"s.download_count desc",
                        "rating":"s.rating_avg desc"
                    }
        order_by = sort_map.get(sort, "s.created_at DESC")
        offset = (page - 1) * size
        params.append(size)
        params.append(offset)
        
                
        sql = f"""
        SELECT
            s.public_id, s.slug, s.display_name, s.summary,
            s.download_count, s.rating_avg, s.created_at,
            c.name AS category_name,
            GROUP_CONCAT(t.name) AS tag_names
        FROM skills s
        LEFT JOIN categories c ON s.category_id = c.id
        LEFT JOIN skill_tags st ON s.id = st.skill_id
        LEFT JOIN tags t ON st.tag_id = t.id
        {where_sql}
        GROUP BY s.id
        order by {order_by}
        LIMIT %s offset %s
        """
        cursor.execute(sql, tuple(params))
        result=cursor.fetchall()
        
        # 把 tag_names 字符串转成列表,也就是我们这个tags
        
        for row in result:
            if row["tag_names"]:
                row["tags"] = row["tag_names"].split(",")
            else:
                row["tags"] = []
            del row["tag_names"]
        return result
    def Find_skill_versions(self,public_id:str):
            cursor=self.conn.cursor()
            try:
                sql="""
select
            skills.display_name AS skill_name,
            jne.name AS founder_name,
            sv.id,
            sv.skill_id,
            sv.version,
            sv.size_bytes,
            sv.extract_file_bytes,
            sv.file_name,
            sv.created_at
 from skills
left join join_and_enter jne on jne.id=skills.owner_id
inner join skill_versions sv on sv.skill_id=skills.id
where skills.public_id=%s
and sv.status=1
order by sv.created_at desc
"""
                cursor.execute(sql,(public_id,))
                result=cursor.fetchall()
                return result
            finally:
                   cursor.close()
#================================================================
    def Re_Find_Zip(self,skill_versions_id:int):
           cursor=self.conn.cursor()
           try:
                  sql="""
select storage_key from skill_versions where id=%s
"""                
                  cursor.execute(sql,(skill_versions_id))
                  result=cursor.fetchone()
                  return result
           finally:
                  cursor.close()
#=================================================================
    def Re_Download_zip(self,user_id:int,skill_versions_id:int):
           cursor=self.conn.cursor()
           try:
                  sql="""
insert into users_download(user_id,version_id)
values(%s,%s)
"""
                  cursor.execute(sql,(user_id,skill_versions_id,))
                  self.conn.commit()
           finally:
                  cursor.close()
#=========================================================================
    async def Re_Upload_Zip(self,
                      upload_zip:UploadFile,
                      name:str,
        tags:list[str],
        category:str,
        version: str,
        owner_id:int,#在service里进行token解码
        readme_html: str,
        summary: str,
        slug:str
        ):
           try:
            cursor=self.conn.cursor()
           #1.先对文件进行审核在搞数据库的写入
            result1=await check_file_zip.Check_file_zip(upload_zip)#处理了内存，file是否缺失等问题
           #2.处理解压后的审核工作
            file_content = result1["file_content"]
            result2=await check_files.Check_Files(file_content=file_content,
                      tags=tags,
                      category=category,
                        version=version,
                        summary=summary,
                        slug=slug)
           #1和2都通过的话，我们开始写insert
            
            sql1="""
select * from skills where slug=%s
"""
            cursor.execute(sql1,(slug,))
            result=cursor.fetchone()
            if not result:#没人发布过
                  sql6="""
                  select id from categories where slug=%s
                  """
                  cursor.execute(sql6,(category,))
                  result6=cursor.fetchone()
                  if not result6:#分类不存在时result6为None,直接取["id"]会TypeError
                        raise HTTPException(status_code=400,detail="分类不存在")
                  sql2="""
insert into skills(slug,public_id,display_name,owner_id,summary,readme_html,category_id)values(%s,%s,%s,%s,%s,%s,%s)
"""
                  
                  temp_public=f"tmp-{slug}"#public_id先占位:不能写死1,第二个技能会撞唯一键uk_public_id
                  cursor.execute(sql2,(slug,temp_public,name,owner_id,summary,readme_html,result6["id"]))
                  new_id=cursor.lastrowid#刚插入那行的自增id,取代原来select回查的sql3
                  uuid=f"uuid-{new_id}"
                  cursor.execute("update skills set public_id=%s where id=%s",(uuid,new_id))
                  storage_key=f"skills/{uuid}/{version}/{slug}_v{version}.zip"
                  sql4="""
insert into skill_versions(skill_id,version,size_bytes,extract_file_bytes,file_name,storage_key)
values(%s,%s,%s,%s,%s,%s)
"""
                  file_name=f"{slug}_v{version}.zip"
                  cursor.execute(sql4,(new_id,version,result1["zip_size"],
                  result1["uncompressed_size"],file_name,storage_key,))
                  new_version_id=cursor.lastrowid#skill_versions的自增id,取代回查的sql5
                  placeholders = ",".join(["%s"] * len(tags))
                  sql7=f"""
select id from tags where slug in ({placeholders})
"""
                  cursor.execute(sql7,tags)
                  result7=cursor.fetchall()
                  for tagId in result7:
                         cursor.execute("insert into skill_tags(skill_id,tag_id)values(%s,%s)", (new_id,tagId["id"],))
                       
                  sql8="""
update skills set latest_version_id=%s where slug=%s
"""
                  cursor.execute(sql8,(new_version_id,slug,))
            else:
                  if owner_id!=result["owner_id"]:#错误
                         Ex_upload_file.Ex_user_ids()
                  #更新版本
                  uuid=f"uuid-{result['id']}"
                  storage_key=f"skills/{uuid}/{version}/{slug}_v{version}.zip"
                  sql9="""
                  insert into skill_versions(skill_id,version,size_bytes,extract_file_bytes,file_name,storage_key)
                  values(%s,%s,%s,%s,%s,%s)
                  """
                  file_name=f"{slug}_v{version}.zip"
                  cursor.execute(sql9,(result["id"],version,result1["zip_size"],
                  result1["uncompressed_size"],file_name,storage_key,))
                  new_version_id=cursor.lastrowid#同if分支,取代回查的sql10
                  sql11="""
update skills set latest_version_id=%s where id=%s
"""
                  cursor.execute(sql11,(new_version_id,result["id"],))
            self.conn.commit()
           except HTTPException:
               raise 
           except Exception:
                 self.conn.rollback()
                 check_files.Ex_upload_file.Ex_insert_zip()
           finally: 
                 cursor.close()
           return {
                        "storage_key":storage_key,
                        "file_name": file_name,
                        "file_content":file_content
                 }


     
#====================================================================
    def Re_All_Tags(self):###用接口进行tags的所有返回
          cursor=self.conn.cursor()
          try:
              sql="""
select * from tags
"""
              cursor.execute(sql)
              result=cursor.fetchall()
              return result#name给前端，slug给后端
          finally:
                cursor.close()
#====================================================================
    def Re_All_categories(self):###用接口进行categories的所有返回
          cursor=self.conn.cursor()
          try:
              sql="""
select id,name,slug from categories
"""
              cursor.execute(sql)
              result=cursor.fetchall()
              return result#name给前端，slug给后端
          finally:
                cursor.close()