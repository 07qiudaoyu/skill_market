import pymysql
from db.session import create_connection
conn=create_connection()
cursor=conn.cursor()
"""
每个技能都有：
id：图书馆内部编号（不对外说）
public_id：书的 ISBN 码（对外展示）
owner_id：作者是谁
category_id：放在哪个书区
slug：书在书架上的英文位置名
display_name：书的封面名称
summary：书背面的简介
readme_html：书的目录和正文
download_count：被借过多少次
rating_avg：读者评分
"""

try:
    sql="""
create table skills(
id int unsigned not null auto_increment,
public_id varchar(36) not null,
owner_id int unsigned not null,
category_id int unsigned not null,
slug varchar(100) not null,
display_name varchar(100) not null,
summary varchar(225) default null,
readme_html text default null,
status tinyint default 1,
latest_version_id int unsigned default null,
download_count int unsigned default 0,
rating_count int unsigned default 0,
rating_avg decimal(2,1) DEFAULT 0.0,
created_at datetime default current_timestamp,
updated_at datetime default current_timestamp on update current_timestamp,
deleted_at datetime default null,
primary key(id),
unique key uk_public_id (public_id),
unique key uk_slug (slug),
key idx_owner_id (owner_id),
key idx_category_id (category_id),
key idx_status (status),
key idx_deleted_at (deleted_at),
key idx_rating_avg (rating_avg)
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
