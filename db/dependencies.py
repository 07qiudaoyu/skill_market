from .session import create_connection


def get_db():#自主关闭conn，也就是数据库
    conn = create_connection()
    try:
        yield conn
    finally:
        conn.close()