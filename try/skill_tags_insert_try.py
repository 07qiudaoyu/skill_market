import pymysql
from db.session import create_connection

conn = create_connection()
cursor = conn.cursor()

# 用 slug 定义技能和标签的对应关系
skill_tag_pairs = [
    ("auto_email_sender", "python"),
    ("auto_email_sender", "email"),
    ("daily-report-generator", "python"),
    ("daily-report-generator", "report"),
    ("web-data-crawler", "python"),
    ("web-data-crawler", "crawler"),
]

try:
    for skill_slug, tag_slug in skill_tag_pairs:
        # 查 skill_id
        cursor.execute("SELECT id FROM skills WHERE slug = %s", (skill_slug,))
        skill_result = cursor.fetchone()
        if not skill_result:
            print(f"找不到 skill: {skill_slug}")
            continue
        skill_id = skill_result["id"]

        # 查 tag_id
        cursor.execute("SELECT id FROM tags WHERE slug = %s", (tag_slug,))
        tag_result = cursor.fetchone()
        if not tag_result:
            print(f"找不到 tag: {tag_slug}")
            continue
        tag_id = tag_result["id"]

        # 插入关系，已存在则忽略
        cursor.execute(
            "INSERT IGNORE INTO skill_tags (skill_id, tag_id) VALUES (%s, %s)",
            (skill_id, tag_id)
        )

    conn.commit()
    print("skill_tags 插入完成")
except pymysql.Error as e:
    print(f"数据库错误: {e}")
    conn.rollback()
finally:
    cursor.close()
    conn.close()
