import pymysql
from db.session import create_connection
conn=create_connection()
cursor=conn.cursor()
try:
    sql="""
create table users_download(
id int unsigned not null auto_increment,
version_id int unsigned not null,
user_id int unsigned not null,
primary key(id),
unique key uq_UV(version_id,user_id),
key idx_us(user_id),
key idx_ve(version_id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
"""
    cursor.execute(sql)
    conn.commit()
    print("建表成功")
except pymysql.Error as e:
    print(f"错误{e}")
    conn.rollback()
finally:
    cursor.close()
    conn.close()

