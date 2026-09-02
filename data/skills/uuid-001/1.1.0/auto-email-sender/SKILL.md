---
name: auto-email-sender
description: 根据模板自动发送邮件，支持 SMTP 配置、HTML 模板渲染和附件发送。
license: MIT
compatibility: Python 3.9+; smtplib
---
# Auto Email Sender
1. 读取邮件模板和收件人列表。
2. 通过 SMTP 连接发送邮件。
3. 支持附件发送。
4. 记录发送日志。

## Examples
输入：模板文件 + 收件人 CSV + 附件
输出：批量发送完成（含附件）