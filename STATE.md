# STATE — 当前状态（2026-11-24）

## 目标
技能市场跑通：FastAPI 后端 + MySQL/MinIO + 静态前端 + Go CLI(skillctl)。

## 已完成
- 后端分层：`api/router.py` 汇总 → `modules/{auth,skills}` 的 router/service/repository。
- 下载记账收敛到服务端流尾单步写库；`download_count` 口径 = 去重后的 (user, version)。
- CLI 命令树：`auth{register,login,logout}` / `skill{list,search,install,uninstall,upgrade}` / `version`。
- CLI 全局 flag：`--token` `--server` `--json`（JSON 信封，stdout 只放结果，提示走 stderr）。
- Makefile：`build` `dist` `check` `fmt` `vet` `test` `clean` `dist-clean`。

## 未提交（工作区现状，别丢）
- 已改动：`root.go` `main.go` `Makefile` `README.md` `oneself/{login,logout,version}.go` `skills/{list,search,skill_url,target,uninstall}.go`
- 未跟踪：`oneself/register.go`、`REPORT.md`（CLI 改造记录，1445 行）

## 坑
- 流尾记账依赖请求级 DB 连接还活着；连接被回收就只 `print`，不报错。
- GET 下载带写副作用，幂等靠 `INSERT IGNORE`。
- `modules/go_mod/.tmp-home/.skillctl/config.json` 曾带真实 JWT 入库，已移出跟踪，但 git 历史里仍有。

## 下一步
1. 提交 CLI 改动前先过一遍 `git diff`；`REPORT.md` 是否入库自己定。
2. 用旧 token 那台机器重新 `skillctl auth login`，让历史里的凭据彻底失效。
3. 起后端验一遍：`uvicorn app.main:app --reload`。
