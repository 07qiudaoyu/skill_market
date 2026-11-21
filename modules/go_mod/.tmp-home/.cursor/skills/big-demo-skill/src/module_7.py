"""模块 ��用来把仓库撑大。"""

import json


def load_config(path: str) -> dict:
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)


def process_7(data: list) -> list:
    return [item for item in data if item]
