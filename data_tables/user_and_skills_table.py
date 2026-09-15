import pymysql
from db.session import create_connection
conn=create_connection()
cursor=conn.cursor()

try:
    sql="""
create table user_skills_versions(
id int unsigned not null auto_increment,
user_id int unsigned not null,
skill_id int unsigned not null,
version_id int unsigned not null,
key idx_u(user_id),
key idx_b(skill_id),
primary key(id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
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
    