# Auth 请求/响应 DTO
'''
这里负责：

前端传进来的数据长什么样，后端返回的数据长什么样。

比如：

class RegisterRequest(BaseModel):
    email: EmailStr
    password: str

登录：

class LoginRequest(BaseModel):
    email: EmailStr
    password: str

返回：

class TokenResponse(BaseModel):
    access_token: str
    refresh_token: str

它就是：

API 数据格式定义。'''

#规定注册接口“进来的数据长什么样”。
from pydantic import BaseModel, Field, EmailStr
from datetime import datetime

class RegisterRequest(BaseModel):
    email: EmailStr
    name: str = Field(min_length=2, max_length=20)
    password: str = Field(min_length=8, max_length=32)
class RegisterResponse(BaseModel):
    message: str
#登入接口的数据规定：
class UserRequest(BaseModel):
    email:EmailStr
    password:str
class UserResponse(BaseModel):
    token:str
    message:str
    user_id:int
    name:str

class User_Data_Response(BaseModel):
    message:str
    user_id:int
    name:str
    developer:int
    join_date:datetime
#全用户得数据
class UserDataItem(BaseModel):
    id:int
    name:str
    email:EmailStr
    developer:int
    join_date:datetime

class Users_Datas_Response(BaseModel):
    message:str
    users_data:list[UserDataItem]

