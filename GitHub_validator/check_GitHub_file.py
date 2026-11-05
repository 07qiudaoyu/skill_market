
from app.core.exceptions import Ex_GitHUb_main
#这个我们就用来做GitHub上传的审核吧
def Va_check_main(data):
    #检查是不是含有我们必要的东西
    file_names={item["path"] for item in data["tree"]}
    count1=sum(n in file_names for n in ("skill.json","SKILL.md","README.md"))
    if count1!=3:
        Ex_GitHUb_main.EX_GitHub_main_error_1()
    
