import pymysql
from db.session import create_connection
conn=create_connection()
cursor=conn.cursor()
try:
    sql="""
    create table if not exists skill_tags(
    skill_id int unsigned not null,
    tag_id int unsigned not null,
    primary key(skill_id, tag_id),
    key idx_tag_id_skill_id (tag_id, skill_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
"""
    cursor.execute(sql)
    conn.commit()
    print("建表成功")
except pymysql.Error as e:
    print(f"数据库错误：{e}")
    conn.rollback()
finally:
    cursor.close()
    conn.close()