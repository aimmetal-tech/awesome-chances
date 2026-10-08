# 贡献指南

欢迎参与 awesome-chances。开始贡献前，请阅读根目录的 [AGENTS.md](./AGENTS.md)，以及目标文件所在路径上所有适用的子 AGENTS.md。

## 新功能必须先有 Issue

**开始编写新功能前，必须先创建对应 Issue**，说明以下内容：

- 功能目标：解决什么问题、面向哪些用户。
- 实现范围：本次需要实现的能力及边界。
- 验收标准：如何判断功能已经完成。

创建前先搜索已有 Issue；如果已有相同功能的 Issue，直接复用，并说明计划承担的工作，避免重复创建。

**创建 PR 后，必须在对应 Issue 正文中用 `Related` 附上实际 PR 链接，同时在 PR 描述中用 `Closes` 或 `Related` 引用对应 Issue。**

- 尚未创建 PR 时，在 Issue 的“关联 PR”部分标记“待创建”；创建后替换为实际链接。
- 同一 Issue 对应多个 PR 时，在 Issue 正文中列出全部 PR 链接。
- 新增或调整 PR 时，同步更新对应 Issue 正文，保持双方关联一致。

### 关联关键字

| 写法 | 适用场景 |
| --- | --- |
| `Closes #101` | PR 完成该 Issue 的全部验收标准，合并后应关闭该 Issue。 |
| `Related #101` | PR 仅与该 Issue 相关，或只完成其中一部分工作，暂不关闭该 Issue。 |
| `Related [PR 标题](PR_URL)` | 在 Issue 正文中回链对应 PR；将 `PR_URL` 替换为实际 PR 链接。 |

关联关键字写在 PR 描述正文中，不只写在标题或评论里。涉及多个 Issue 时，每个 Issue 单独写一行。

`Closes` 是 GitHub 支持的自动关闭关键字，PR 需以仓库默认分支为目标，并在合并到默认分支后关闭对应 Issue。`Related` 是本项目约定的普通关联写法，不是自动关闭关键字，也不会自动关闭 Issue。参见 [GitHub 官方说明](https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/linking-a-pull-request-to-an-issue)。

同一 Issue 拆成多个 PR 时，中间 PR 使用 `Related`；完成全部验收标准的最后一个 PR 才使用 `Closes`。

## 大 Issue 与子 Issue

贡献者可以使用**一个大 Issue 和多个子 Issue**拆分新功能。

- 大 Issue 记录整体目标、总体验收标准和子 Issue 清单，用于跟踪整个功能。
- 每个子 Issue 记录具体任务、实现范围、验收标准、父 Issue 链接及对应 PR 链接。
- 父 Issue 列出子 Issue，子 Issue 引用父 Issue，保持双向关联。
- 每个实现 PR 使用 `Closes` 引用已全部完成的子 Issue，使用 `Related` 引用父 Issue 或尚未完成的子 Issue；子 Issue 正文必须用 `Related` 补充对应 PR 链接。
- 大 Issue 汇总该功能对应的全部 PR 链接；完成一个子任务时，更新父 Issue 的任务进度。
- 只完成部分子任务的 PR 不得使用 `Closes` 关闭父 Issue；父 Issue 的全部验收标准完成后才可关闭。

## 填写示例

以下编号及 `PR_URL` 均为示例占位，使用时替换为真实的 Issue 编号和 PR URL。示例中的验收项是功能目标，不代表当前项目已经实现这些能力。

### 普通功能 Issue

```markdown
## 功能目标
让学生查看并编辑自己的兴趣方向。

## 实现范围
实现兴趣编辑页面、BFF 接口及后端保存逻辑。

## 验收标准
- [ ] 用户可以新增、修改和删除兴趣方向。
- [ ] 保存后刷新页面仍能看到修改结果。
- [ ] 保存失败时显示中文提示。

## 关联 PR
- 待创建
```

创建 PR 后，将“关联 PR”中的“待创建”替换为实际链接；多个 PR 分别列出：

```markdown
## 关联 PR
- Related [兴趣编辑功能 PR](PR_URL)
```

### 大 Issue（示例：#100）

```markdown
## 整体目标
完成用户兴趣画像的编辑与保存。

## 总体验收标准
- [ ] 页面、BFF 与后端接口能够完成兴趣编辑流程。
- [ ] 数据保存成功，错误提示清晰。

## 子 Issue
- [ ] #101 兴趣编辑页面与 BFF
- [ ] #102 后端兴趣保存接口

## 关联 PR
- 待创建
```

### 子 Issue（示例：#101）

```markdown
## 父 Issue
#100

## 具体任务与范围
实现兴趣编辑页面与 Next.js BFF，调用后端兴趣保存接口。

## 验收标准
- [ ] 页面展示和编辑兴趣方向。
- [ ] BFF 正确转发请求并适配页面数据。
- [ ] 请求失败时展示中文提示。

## 关联 PR
- 待创建
```

### PR 描述

```markdown
## 关联 Issue
Closes #101
Related #100

## 改动说明
实现兴趣编辑页面与 BFF 接口。

## 验证结果
- pnpm lint：通过。
- pnpm build：通过。
- 手动验证：兴趣编辑成功和后端请求失败的场景。

## 文档同步
已更新受影响的文档；无结构性变更时注明“不适用”。
```

以上示例假设 PR 已完成子 Issue #101 的全部验收标准；若仅完成部分工作，将 `Closes #101` 改为 `Related #101`。

提交时按实际执行的检查填写验证结果；未执行的检查应说明原因，不能照抄示例中的“通过”。PR 创建后，用 `Related [PR 标题](实际 PR URL)` 补充到对应 Issue 正文；采用父子 Issue 时，还需更新父 Issue 的 PR 汇总。

## 开发与提交检查

遵循 [README.md](./README.md) 中的本地开发说明，并读取各级适用的 AGENTS.md。前端代码放在 `src/`，BFF 由 Next.js 在 `src/app/api/` 实现，后端服务代码放在 `backend/`。

界面和文档默认使用简体中文，当前不引入 i18n。中文字体遵循项目的思源黑体优先及系统字体回退约定。

提交 PR 前检查：

- [ ] 新功能已经创建或关联对应 Issue，并写明目标、范围与验收标准。
- [ ] PR 描述使用 `Closes` 或 `Related` 引用对应 Issue，Issue 正文已使用 `Related` 附上全部对应 PR 链接。
- [ ] 使用 `Closes` 的 Issue 已满足全部验收标准；仅部分完成的 Issue 和尚未完成的父 Issue 使用 `Related`。
- [ ] 使用父子 Issue 时，父子关联、任务进度及父 Issue 的 PR 汇总已同步。
- [ ] 代码贡献已运行适用的检查，并在 PR 中记录结果；Next.js 代码按改动运行 `pnpm lint`、`pnpm build` 及相关场景验证，后端代码遵循适用的子 AGENTS.md。
- [ ] 结构性变更已同步更新根 AGENTS.md、README.md 和受影响的子 AGENTS.md；贡献流程变更还须同步更新本指南。
- [ ] 仅修改文档时，已检查 Markdown、链接及空白错误，无需运行生产构建。
