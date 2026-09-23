import asyncio
from starlette.datastructures import UploadFile
from validator.check_file_zip import Check_file_zip

f = open("data/skills/uuid-001/1.0.0/auto-email-sender_v1.0.0.zip", "rb")
uf = UploadFile(filename="a.zip", file=f)

print(asyncio.run(Check_file_zip(uf)))
