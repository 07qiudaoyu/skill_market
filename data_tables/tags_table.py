import pymysql
from db.session import create_connection
conn=create_connection()
cursor=conn.cursor()
"""
为什么需要单独一张表？因为：一本书可以有多个标签："""
try:
    sql="""
create table if not exists tags(
id int unsigned not null auto_increment,
slug varchar(100) not null unique,
name varchar(50) not null,
primary key(id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
"""
    cursor.execute(sql)
    conn.commit()
except pymysql.Error as e:
    print(f"数据库错误：{e}")
    conn.rollback()
finally:
    cursor.close()
    conn.close()
