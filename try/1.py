import re
import requests
import json
url = "https://github.com/07qiudaoyu/skill_market"

pattern = r"https?://github\.com/([^/]+)/([^/]+?)(?:\.git)?/?"#问号这个我没看懂

result = re.fullmatch(pattern, url)#这个咋用的

if not result:
    raise ValueError("GitHub URL 格式不正确")

owner = result.group(1)
repo = result.group(2)


api_url = f"https://api.github.com/repos/{owner}/{repo}"

r = requests.get(api_url, timeout=10)

if r.status_code == 404:
    raise ValueError("GitHub 仓库不存在或无法访问")

r.raise_for_status()

data = r.json()

print(data["full_name"])
print(data["html_url"])
print(data["private"])

print(data)
