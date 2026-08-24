import pymysql
from db.session import create_connection
conn=create_connection()
cursor=conn.cursor()
try:
    sql="""
insert into skills(
public_id,owner_id,category_id,slug,display_name,summary,readme_html,status,
latest_version_id,download_count,rating_count,rating_avg
)values(
'uuid-001',1,1,'auto_email_sender','邮件自动发送工具',
'跟据模版来自动发送件',
'<h1>自动发邮件工具</h1><p>详细说明...</p>',
1, NULL, 12, 3, 4.5
),
(
    'uuid-002', 1, 3, 'daily-report-generator', '日报自动生成器',
    '自动读取数据并生成 Excel/PDF 日报',
    '<h1>日报自动生成器</h1><p>详细说明...</p>',
    1, NULL, 8, 2, 4.0
),
(
    'uuid-003', 1, 1, 'web-data-crawler', '网页数据采集器',
    '输入 URL 自动抓取网页内容并导出',
    '<h1>网页数据采集器</h1><p>详细说明...</p>',
    1, NULL, 25, 5, 4.8
);
"""
    cursor.execute(sql)
    conn.commit()
except pymysql.Error as e:
    print(f"数据库错误{e}")
    conn.rollback()
finally:
    conn.close()
    cursor.close()


    