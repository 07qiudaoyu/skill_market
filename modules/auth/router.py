'''接收 HTTP 请求，然后调用 Service。

不要在这里写大量 SQL。'''

from fastapi import APIRouter, Depends, status

from .schemas import (
    RegisterRequest,
    RegisterResponse,
    UserRequest,
    UserResponse,
    
   
    Users_Datas_Response
    )
from app.core.middleware import header_token
from .service import AuthService
from .dependencies import get_auth_service

router=APIRouter(
    prefix="/auth",
    tags={"Auth"}
)
@router.post(
    "/register",
    response_model=RegisterResponse,
    status_code=status.HTTP_201_CREATED
    )
def register(
    item:RegisterRequest,
    service:AuthService=Depends(get_auth_service)
):
    service.register(
        email=item.email,
        name=item.name,
        password=item.password
    )
    return RegisterResponse(
        message="注册成功"
    )

@router.post(
    "/login",
    response_model=UserResponse,
    status_code=status.HTTP_200_OK
)
def login(
        item:UserRequest,
        service:AuthService=Depends(get_auth_service)
):
    user = service.login(
        email=item.email,
        password=item.password
    )
    token=service.get_token(user)

    return {
        "token":token,
        "message": "登录成功",
        "user_id": user["id"],
        "name": user["name"]
    }#----------------------------------------------------------
@router.post(
    "/user_data",
   
    status_code=status.HTTP_200_OK
)
def data_1_user(
    payload:dict=Depends(header_token),
    service:AuthService=Depends(get_auth_service)
):
    payload=payload
    TOKEN=service.analyze_token(payload)

    return {
        "message": "个人用户数据查询成功",
        "user_id":TOKEN["id"],
        "name":TOKEN["name"],
        "developer":TOKEN["developer"],
        "join_date":TOKEN["join_date"]
        
    }#-----------------------------------------------------------------
@router.post(
    "/users_datas",
   
    response_model=Users_Datas_Response
)
def data_2_user(
    payload: dict = Depends(header_token),
     service:AuthService=Depends(get_auth_service)
):
    TOKEN=service.analyze_token_pro(
        payload
    )
    return {
        "message":"全用户数据查询成功",
        "users_data":TOKEN
    }
    
