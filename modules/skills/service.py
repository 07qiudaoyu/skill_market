from .repository import SkillsRepository

class SkillsService:
    def __init__(self, repository: SkillsRepository):
        self.repository = repository

    def search_skills(self, q=None, category=None, tags=None, sort="newest", page=1, size=10):
        items = self.repository.select_skills_re(
            q=q, category=category, tags=tags,
            sort=sort, page=page, size=size
        )
        return {
            "message": "搜索成功",
            "items": items,
            "page": page,
            "size": size
        }