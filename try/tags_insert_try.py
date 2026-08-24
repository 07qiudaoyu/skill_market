import pymysql
from db.session import create_connection
conn=create_connection()
cursor=conn.cursor()
try:
    sql="""
insert into tags(
slug,name
)values
('python', 'Python'),
('javascript', 'JavaScript'),
('email', '邮件'),
('crawler', '爬虫'),
('api', 'API'),
('ai', '人工智能'),
('webhook', 'Webhook'),
('report', '报表'),
('excel', 'Excel'),
('pdf', 'PDF');
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