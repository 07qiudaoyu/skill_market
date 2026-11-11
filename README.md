# Skill Market

一个给 agent 用的技能市场。

用户把自己的技能打成包丢上来，或者直接给一个 GitHub 仓库地址。平台审核通过后入库，别人能搜到、能下载、能装给自己的 agent 用。

后端 FastAPI + MySQL，前端是原生 HTML/JS，另有一个 Go 写的命令行客户端。

## 功能

- 注册、登录，JWT 鉴权
- 技能搜索，按分类、标签筛选，分页
- 上传 zip：校验根目录三件套、字段一致性、大小限制
- 从 GitHub 仓库导入：只收仓库地址，后端自动取默认分支、拉包、剥离顶层目录、审核入库
- 下载技能 zip，并记录下载次数
- 按技能查全部版本
- 分类和标签列表，供前端做筛选

## 技术栈

| 层 | 用什么 |
| --- | --- |
| 后端 | FastAPI、pymysql、PyJWT、bcrypt |
| 数据库 | MySQL |
| 前端 | 原生 HTML / CSS / JS，由 FastAPI 挂在根路径 |
| CLI | Go |

## 目录结构

```text
skill_market/
├── app/                 FastAPI 实例与内核（配置、异常、JWT、中间件）
├── api/                 路由汇总，统一前缀 /api/v1
├── modules/
│   ├── auth/            注册、登录、用户数据
│   ├── skills/          搜索、上传、下载、版本、GitHub 导入
│   └── go_mod/          Go 命令行客户端
├── validator/           zip 包审核（大小、根目录、字段一致性）
├── GitHub_validator/    GitHub 接口审核（状态码、目录）
├── db/                  数据库连接
├── data_tables/         建表脚本
├── data/skills/         技能 zip 落盘目录，按 uuid/版本/名字
├── web/                 前端页面
├── try/                 一次性脚本，试 SQL 用的
├── requirements.txt
└── .env.example
```

每个业务模块都分四层：

```text
router.py      收请求、做参数校验
service.py     业务逻辑，拼字段、调审核
repository.py  只写 SQL
schemas.py     请求和响应的数据结构
```

## 快速开始

### 1. 建库

```sql
CREATE DATABASE skill_market DEFAULT CHARSET utf8mb4;
```

### 2. 配置环境变量

复制 `.env.example` 为 `.env`，按自己的环境改：

```text
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=你的密码
DB_NAME=skill_market

JWT_SECRET_KEY=换成一个长一点的随机串
GITHUB_TOKEN=你的 GitHub token
```

`GITHUB_TOKEN` 用来调 GitHub 接口。不带的话匿名只有 60 次一小时，带了是 5000 次。

### 3. 装依赖

```bash
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

注意：`requirements.txt` 里有 `re` 和 `jsona` 两行，前者是 Python 自带的、后者装不上，安装前先删掉这两行。

### 4. 建表

依次执行 `data_tables/` 下的脚本，顺序按外键依赖来：

```text
join_and_enter_table.py   用户
categories_table.py       分类
tags_table.py             标签
skills_table.py           技能
skill_versions_table.py   版本
skill_tags_table.py       技能和标签的关系
user_and_skills_table.py
users_download_table.py
```

### 5. 启动

```bash
uvicorn app.main:app --reload --port 8000
```

打开 `http://127.0.0.1:8000`，就是前端页面。

接口文档在 `http://127.0.0.1:8000/docs`。

## 接口一览

统一前缀 `/api/v1`。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/auth/register` | 注册 |
| POST | `/auth/login` | 登录，返回 JWT |
| POST | `/auth/user_data` | 查当前用户，需要 JWT |
| GET | `/skills/search` | 搜索，支持 q、category、tags、sort、page、size |
| GET | `/skills/find_versions` | 查某个技能的全部版本 |
| GET | `/skills/Download_file` | 下载 zip，需要 JWT |
| POST | `/skills/upload_zip` | 上传 zip，需要 JWT |
| POST | `/skills/GitHub_post` | 从 GitHub 仓库导入，需要 JWT |
| GET | `/skills/All_Tags` | 所有标签 |
| GET | `/skills/All_categories` | 所有分类 |

## 上传规范

### 文件

zip 根目录必须包含三个文件：

```text
skill.json
SKILL.md
README.md
```

放在子文件夹里审核不通过。

### 大小

```text
zip 包            ≤ 50MB
解压后内容合计    ≤ 100MB
```

### skill.json

表单里填的 slug、版本、分类、标签、简介，必须和 `skill.json` 里的 `name`、`version`、`category`、`tags`、`description` 逐项一致。

标签不看顺序，看集合。

### SKILL.md

文件必须以 `---` 开头，frontmatter 里：

- `name` 与 slug 一致
- `description` 与简介一致
- `name` 长度 1 到 64
- `description` 长度 1 到 1024

文件必须是 UTF-8。

### 分类和标签

只能取库里已有的值，写进 `skill.json` 的是小写 slug。

## 从 GitHub 导入

只收仓库根地址：

```text
https://github.com/{owner}/{repo}
```

带 `/tree/main` 之类的地址会被拒。

后端流程：

```text
地址 → 校验格式 → 拉 zipball → 剥掉顶层目录 → 根目录审核
     → 从 skill.json 取元数据 → 校验字段 → 入库 → zip 落盘
```

分支不用用户给，后端自己读仓库的 `default_branch`。

## 命令行客户端

在 `modules/go_mod/` 下。

```bash
go run main.go rg <email> <name> <password>
go run main.go lg <email> <password>
go run main.go fv <public_id>
go run main.go dl <version_id>
```

token 存在系统 keyring 里，登录一次就行。

## 11.11进行了优化cli，采用cobra包，后期将对输出格式进行优化及keyring缓存优化，目前keyring不具备删除缓存能力
