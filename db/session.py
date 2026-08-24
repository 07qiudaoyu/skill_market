'''负责：

怎么连接 MySQL。'''
import pymysql


def create_connection():
    return pymysql.connect(
        host="localhost",
        user="root",
        password="20070511@Xx",
        database="skill_market",
        cursorclass=pymysql.cursors.DictCursor
    )