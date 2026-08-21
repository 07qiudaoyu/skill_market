'''这个文件解决：

每个 HTTP 请求怎么获得数据库 Session。
auth/dependencies.py

这是 Auth 模块内部使用的依赖。

例如：

创建 AuthService
创建 AuthRepository

也就是说：

FastAPI Depends
      ↓
AuthService
      ↓
AuthRepository'''
from fastapi import Depends

from db.dependencies import get_db
from .repository import AuthRepository
from .service import AuthService


def get_auth_service(
    conn=Depends(get_db)
):
    repository = AuthRepository(conn)
    return AuthService(repository)