import pymysql
from db.session import create_connection
conn=create_connection()
cursor=conn.cursor()
try:
    sql="""
INSERT INTO skill_versions (skill_id, version, size_bytes, file_name, storage_key, status) VALUES
(1, '1.0.0', 102400, 'auto_email_sender_v1.0.0.zip', 'skills/uuid-001/1.0.0/auto_email_sender_v1.0.0.zip', 1),
(1, '1.1.0', 153600, 'auto_email_sender_v1.1.0.zip', 'skills/uuid-001/1.1.0/auto_email_sender_v1.1.0.zip', 1),
(2, '1.0.0', 204800, 'daily_report_generator_v1.0.0.zip', 'skills/uuid-002/1.0.0/daily_report_generator_v1.0.0.zip', 1),
(3, '1.0.0', 307200, 'web_data_crawler_v1.0.0.zip', 'skills/uuid-003/1.0.0/web_data_crawler_v1.0.0.zip', 1);
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