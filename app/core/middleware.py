#“所有请求都会经过的门卫”
from fastapi.security import HTTPBearer
from fastapi.security import HTTPAuthorizationCredentials
from fastapi import Depends
from app.core.security import check_token
from fastapi import HTTPException
Token=HTTPBearer()#拿到token
async def header_token(
credentials:HTTPAuthorizationCredentials = Depends(Token)#单从后端里面拿，不走前端
        
):
    token=credentials.credentials#拿到前端的token
    payload=check_token(token)
    if not payload:
        raise HTTPException(
            status_code=401,
            detail="token错误，header出错"
        )
    return payload
