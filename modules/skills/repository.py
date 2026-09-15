
import uuid


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
                  
    def Re_Upload_Zip(
        self,
        name: str,
        version: str,
        user_id: int,
        category: str,
        change_tags: list[str],
        readme_html: str,
        summary: str,
        slug: str,
        size_bytes: int
):
     cursor = self.conn.cursor()

     try:
        sql1 = """
        SELECT *
        FROM skills
        WHERE slug = %s AND owner_id = %s
        """

        cursor.execute(sql1, (slug, user_id))
        result1 = cursor.fetchone()

        if not result1:  # 新技能

            sql2 = """
            SELECT *
            FROM categories
            WHERE slug = %s
            """

            cursor.execute(sql2, (category,))
            result2 = cursor.fetchone()

            if not result2:
                raise ValueError("分类不存在")

            sql3 = """
            INSERT INTO skills
                (public_id, display_name, summary, readme_html, slug, owner_id, category_id)
            VALUES
                (%s, %s, %s, %s, %s, %s, %s)
            """

            cursor.execute(
                sql3,
                (
                    str(uuid.uuid4()),
                    name,
                    summary,
                    readme_html,
                    slug,
                    user_id,
                    result2["id"]
                )
            )

            # 获取刚刚插入的 skill_id
            skill_id = cursor.lastrowid

            # 处理标签
            if change_tags:
                placeholders = ",".join(
                    ["%s"] * len(change_tags)
                )

                sql5 = f"""
                SELECT id
                FROM tags
                WHERE slug IN ({placeholders})
                """

                cursor.execute(sql5, change_tags)
                result5 = cursor.fetchall()

                tag_ids = [
                    row["id"]
                    for row in result5
                ]

                for tag_id in tag_ids:
                    sql_tag = """
                    INSERT INTO skill_tags
                        (skill_id, tag_id)
                    VALUES
                        (%s, %s)
                    """

                    cursor.execute(
                        sql_tag,
                        (skill_id, tag_id)
                    )

            # 插入版本
            public_id = f"uuid-{skill_id}"
            file_name = f"{slug}_v{version}.zip"
            storage_key = f"skills/{public_id}/{version}/{file_name}"

            sql6 = """
            INSERT INTO skill_versions
                (version, file_name, skill_id, storage_key, size_bytes)
            VALUES
                (%s, %s, %s, %s, %s)
            """

            cursor.execute(
                sql6,
                (
                    version,
                    file_name,
                    skill_id,
                    storage_key,
                    size_bytes
                )
            )

            version_id = cursor.lastrowid

            sql8 = """
            UPDATE skills
            SET public_id = %s,
                latest_version_id = %s
            WHERE id = %s
            """

            cursor.execute(
                sql8,
                (
                    public_id,
                    version_id,
                    skill_id
                )
            )

        else:  # 已存在 Skill

            skill_id = result1["id"]
            public_id = result1["public_id"]

            file_name = f"{slug}_v{version}.zip"
            storage_key = f"skills/{public_id}/{version}/{file_name}"

            sql9 = """
            INSERT INTO skill_versions
                (version, file_name, skill_id, storage_key, size_bytes)
            VALUES
                (%s, %s, %s, %s, %s)
            """

            cursor.execute(
                sql9,
                (
                    version,
                    file_name,
                    skill_id,
                    storage_key,
                    size_bytes
                )
            )

            version_id = cursor.lastrowid

            sql11 = """
            UPDATE skills
            SET summary = %s,
                readme_html = %s,
                latest_version_id = %s
            WHERE id = %s
            """

            cursor.execute(
                sql11,
                (
                    summary,
                    readme_html,
                    version_id,
                    skill_id
                )
            )

        self.conn.commit()

        return {
            "file_name": file_name,
            "storage_key": storage_key
        }

     except Exception:
        self.conn.rollback()
        raise

     finally:
        cursor.close()