# User/Session 查询与写入
'''这个文件负责：

和数据库打交道。

例如：

get_user_by_email()
get_user_by_id()
create_user()
create_refresh_session()
revoke_refresh_session()

Service：

“我要找这个邮箱的用户”

Repository：

“好的，我去 MySQL 查”

所以你可以记：

Service 决定做什么，Repository 决定怎么查数据库。'''

class AuthRepository:
    def __init__(self,conn):
        self.conn=conn
    def get_user_by_email(self,email:str)->dict | None:
        cursor=self.conn.cursor()
        cursor.execute(
            "select * from join_and_enter where email=%s",
            (email,)
        )
        result=cursor.fetchone()
        
        return result
        

    def email_exists(self,email:str) ->bool:
        cursor=self.conn.cursor()

        cursor.execute(
            "select count(*) as count from join_and_enter where email=%s",
            (email,)
        )

        result=cursor.fetchone()
        return result["count"]>0
    
    def create_user(
            self,
            email:str,
            name:str,
            password_hash:str
    ):
        cursor=self.conn.cursor()
        cursor.execute(
            "insert into join_and_enter(email,name,password)values(%s,%s,%s)",
            (email,name,password_hash,)
        )
        self.conn.commit()


    def data_user(self,email:str):
        cursor=self.conn.cursor()
        cursor.execute(
            "select id,name,email,developer,join_date from join_and_enter where email=%s",
            (email,)
        )
        result=cursor.fetchone()
        return result
    def get_all_data(self):
        cursor=self.conn.cursor()
        cursor.execute(
            "select id,name,email,developer,join_date from join_and_enter"
        )
        result1=cursor.fetchall()
        return result1
    
    

        