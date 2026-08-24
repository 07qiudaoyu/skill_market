import pymysql
from db.session import create_connection
conn=create_connection()
cursor=conn.cursor()
try:
    sql="""
insert into categories(
slug,name,description,sort_order,is_active
)values
('web-dev', 'Web开发', '前端、后端、全栈相关技能', 1, 1),
('data-ai', '数据与AI', '数据分析、机器学习、人工智能相关', 2, 1),
('automation', '自动化工具', '自动化办公、定时任务、流程自动化', 3, 1),
('productivity', '效率工具', '提升工作效率的小工具', 4, 1),
('devops', '运维与部署', '服务器、部署、监控相关技能', 5, 1);
"""
    cursor.execute(sql)
    conn.commit()
except pymysql.Error as e:
    print(f"数据库错误{e}")
    conn.rollback()
finally:
    print("加入成功")
    cursor.close()
    conn.close()

