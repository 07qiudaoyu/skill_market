---
name: root-demo-skill
description: "用来测试 GitHub 仓库审核的最小技能包，三件套都在根目录。"
license: MIT
compatibility: "Python 3.9+"
---

# Root Demo Skill

一个只用来测试的仓库。

根目录固定放三样东西：

- `skill.json`
- `SKILL.md`
- `README.md`

`examples/` 是故意留的子目录，用来对比 `recursive=1` 前后的返回。

## 用法

```bash
python examples/hello.py
```
