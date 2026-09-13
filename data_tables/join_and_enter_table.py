import pymysql
from db.session import create_connection

conn = create_connection()
cursor = conn.cursor()

try:
    sql = """
create table join_and_enter(
    id int unsigned not null auto_increment,
    name varchar(225) not null,
    email varchar(225) not null,
    password varchar(225) not null,
    developer tinyint unsigned default 0,
    status tinyint unsigned default 1,
    email_legal tinyint unsigned default 1,
    join_date datetime default current_timestamp,
    update_at datetime default current_timestamp on update current_timestamp,
    deleted_at datetime default null,
    primary key (id),
    unique key uk_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
"""
    cursor.execute(sql)
    conn.commit()
    print("建表成功")
except pymysql.Error as e:
    print(f"错误: {e}")
    conn.rollback()
finally:
    cursor.close()
    conn.close()
