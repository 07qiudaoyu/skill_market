from zipfile import ZipFile
with ZipFile("data/skills/uuid-001/1.0.0/auto-email-sender_v1.0.0.zip") as z:
    for n in z.namelist():
        print(repr(n))
