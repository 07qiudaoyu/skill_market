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