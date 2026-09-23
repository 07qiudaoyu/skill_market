import pymysql
from db.session import create_connection
conn=create_connection()
cursor=conn.cursor()
try:
    sql="""
create table skill_versions(
id int unsigned not null auto_increment,
skill_id int unsigned not null,
version varchar(50) not null,
size_bytes int unsigned default 0,
extract_file_bytes int unsigned default 0,
file_name varchar(500) not null,
storage_key varchar(500) not null,
status tinyint default 1,
created_at datetime default current_timestamp,
primary key (id),
unique key uk_skill_version (skill_id, version),
unique key uk_storage_key (storage_key)
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


