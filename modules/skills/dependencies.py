from fastapi import Depends

from db.dependencies import get_db
from .repository import SkillsRepository
from .service import SkillsService


def get_skill_service(
    conn=Depends(get_db)
):
    repository = SkillsRepository(conn)
    return SkillsService(repository)