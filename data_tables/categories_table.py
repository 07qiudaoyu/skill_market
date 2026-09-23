
import pymysql
from db.session import create_connection
conn=create_connection()
cursor=conn.cursor()

try:
    sql="""
create table if not exists categories(
id int unsigned not null auto_increment,
slug varchar(50) not null,
name varchar(50) not null,
description varchar(225) default null,
sort_order int default 0,
is_active tinyint default 1,
created_at datetime default current_timestamp,
updated_at datetime default current_timestamp on update current_timestamp,
primary key (id),
unique key uk_slug (slug),
key idx_is_active_sort_order (is_active,sort_order)
)
ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='技能分类表';
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

