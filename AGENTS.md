# skill_market —— 项目上下文（给 agent）

技能市场：FastAPI 后端 + MySQL + 静态前端 + Go CLI。主人用 Qoder 做规划/前端；我负责监督、答疑、按需改后端。

## 开工顺序
1. 先读同目录 `STATE.md`（≤30 行：当前目标、已完成、坑、下一步）。
2. 只在缺关键事实时问一句，别让主人复述历史。
3. 回复默认短；代码判断点到为止，别列长清单、别倒整段代码。

## 运行
- 后端：`uvicorn app.main:app --reload`（8000）
- MySQL：`skill_market` 库，连接参数在 `.env`
- Go CLI：`cd modules/go_mod && go run . <cmd>`
- 前端：`web/` 静态页，挂在 `/`

## 结构
- `app/` 装配与 core（config/security/middleware/异常/分页）
- `api/router.py` 汇总路由
- `modules/<域>/`：router → service → repository + schemas/dependencies
- `data_tables/` 建表/表操作脚本；`db/` 连接与 MinIO
- `modules/go_mod/` Go CLI；`web/` 前端；`data/` 上传的 zip
- `try/`、`validator/` 是草稿/校验，别当主链路

## 已定约定
- 分层：路由不写业务，service 不写 SQL，SQL 只在 repository。
- 一次业务动作 = 一个事务 = 一次 commit。
- 下载记账在服务端流尾部完成；`/Updata_Download` 和 CLI 第二步已删。
- `download_count` 现口径 = 去重后的 (user, version) 组合，不是总下载次数。
- `web/` 前端交给 Qoder，别主动改。

## 已知坑
- 流尾记账依赖请求级 DB 连接还活着；连接被回收就只 `print`，不报错。
- GET 下载带写副作用，语义不干净（幂等靠 `INSERT IGNORE`）。
- 旧 `README.md` 是一份 Qoder 审查记录，不是项目说明，别全信。
