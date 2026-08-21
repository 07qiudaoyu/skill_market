# Argon2、JWT、refresh token
#安全部门
'''这个文件专门负责：
只负责“技术上的安全操作”
密码、JWT、Refresh Token 这些安全相关的东西。'''

import bcrypt
import jwt
def hash_password(password: str) -> str:
    password_hash = bcrypt.hashpw(
        password.encode("utf-8"),
        bcrypt.gensalt()
    )
    return password_hash.decode("utf-8")
#哈希加密文件代码｜^
#对比加密密码代码||
def make_hash_password(password: str,hash_password:str)->bool:
    return bcrypt.checkpw(
        password.encode("utf-8"),
        hash_password.encode("utf-8")
    )
#token编写
SECRET_KEY="abuiwhfbiujeuihd133"
def token_jwt(payload:dict):
    token=jwt.encode(
        payload,
        SECRET_KEY,
        algorithm="HS256"
    )
    return token
#token的解码
def check_token(token:str):
    try:
        payload1=jwt.decode(
            token,
            SECRET_KEY,
            algorithms=["HS256"]
        )
    except jwt.InvalidTokenError:
        return None
    return payload1


