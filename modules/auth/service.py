'''这是非常重要的。

它负责：

真正的登录/注册业务逻辑。

例如注册：

用户提交注册
 ↓
检查邮箱是否存在
 ↓
检查用户状态
 ↓
密码 Argon2
 ↓
创建 User
 ↓
保存数据库

登录：

email
 ↓
查用户
 ↓
验证密码
 ↓
检查用户状态
 ↓
生成 JWT
 ↓
创建 RefreshSession
 ↓
返回 Token

所以：

业务逻辑放 service。'''
from fastapi import HTTPException
from app.core.security import hash_password, make_hash_password,token_jwt,check_token
from .repository import AuthRepository
import time

class AuthService:
    def __init__(self,repository:AuthRepository):
        self.repository=repository
    def register(
            self,
            email:str,
            name:str,
            password:str
    ):
        if self.repository.email_exists(email):
            raise HTTPException(
                status_code=409,
                detail="邮件已被注册"
            )
        password_hash=hash_password(password)
        self.repository.create_user(
            email=email,
            name=name,
            password_hash=password_hash
        )
    def login(
            self,
            email:str,
            password:str
    ):
        user=self.repository.get_user_by_email(email)
        if not user:
            raise HTTPException(
                status_code=401,
                detail="邮件或密码错误"
            )
        else:
            password_check=make_hash_password(
                password,
                user["password"]
            )
            if not password_check:
                raise HTTPException(
                status_code=401,
                detail="邮箱或密码错误")
            return user
    #生成token
    def get_token(self, user:dict):
        payload={
                "id":user["id"],
                "developer":user["developer"],
                "email":user["email"],
                "exp":int(time.time())+60*60*24
            }
        token=token_jwt(payload)
        return token
    #解码token
    def analyze_token(self,token:str):
        result=check_token(token)
        if result is None:
            raise HTTPException(
                status_code=401,
                detail="token错误!"
            )
        else:
             ak=result
             UserData=self.repository.data_user(ak["email"])
             return UserData
             



    
            