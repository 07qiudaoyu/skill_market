# skill_demo_ok

测试仓库，用来验证「GitHub 仓库根目录是否包含三件套」。

## 结构

```text
/
├── skill.json
├── SKILL.md
├── README.md
├── examples/
│   ├── hello.py
│   └── data/
│       └── note.txt
└── docs/
    └── usage.md
```

根目录 5 项：3 个文件 + 2 个目录。

不加 `recursive`，`/git/trees/main` 只回这 5 条。

加了 `?recursive=1`，一共 9 条，`examples/data/note.txt` 这种也会出来。
