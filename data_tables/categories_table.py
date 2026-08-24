"""表 1：categories（分类表）
这张表是什么： 就像淘宝的商品分类（手机、电脑、衣服），技能市场也需要给技能分类，方便用户按分类浏览。
字段	类型	解释
id	int	主键，自动递增。每个分类的唯一编号
slug	varchar	URL 友好的英文标识，比如 "web-dev"、"machine-learning"。用在网址里，比中文好看
name	varchar	分类的显示名称，比如 "Web开发"、"机器学习"
description	varchar	分类的描述，比如 "前端后端相关技能"
sort_order	int	排序权重。数字越小排越前面，用来控制分类在页面上的显示顺序
is_active	tinyint	是否启用。1 = 显示，0 = 隐藏。不想删除分类，但暂时不想让用户看到时用
1. skill_categories（分类 = 书区）
web-dev    → Web开发
data-ai    → 数据与AI
automation → 自动化工具
作用： 用户进图书馆，先选"我想去科技区"。分类让东西有层级，
一个技能只能属于一个分类。为什么只有一个分类？ 一本书不会同时放在两个书区。技能同理。
"""

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

