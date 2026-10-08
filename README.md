# awesome-chances

使用 `pnpm create next-app` 初始化的 Next.js 项目，采用 TypeScript、App Router、Tailwind CSS 和 ESLint。

## 本地开发

```bash
pnpm install
pnpm dev
```

浏览器打开 [http://localhost:3000](http://localhost:3000)。修改 `src/app/page.tsx` 后，页面会自动更新。

## 检查与构建

```bash
pnpm lint
pnpm build
pnpm start
```

`pnpm start` 启动生产服务，需要先运行 `pnpm build`。

## 项目结构

- `backend/`：后端服务代码目录，包含服务接口、核心业务逻辑、Agent 编排和数据访问。
- `src/app/api/`：Next.js BFF 接口的约定目录，使用 Route Handlers 转发请求、聚合接口并适配页面数据；当前尚未实现。
- `src/app/page.tsx`：首页。
- `src/app/layout.tsx`：根布局与页面元数据。
- `src/app/globals.css`：全局样式与 Tailwind CSS。
- `public/`：静态资源。
- `next.config.ts`：Next.js 配置。
- `pnpm-lock.yaml`：依赖锁文件。

BFF 层由 Next.js 实现，通过后端接口调用 `backend/` 服务；核心业务逻辑、Agent 编排和数据访问统一在 `backend/` 实现。

导入别名 `@/*` 指向 `src/*`。

## 语言与字体

项目默认使用简体中文，页面语言为 `zh-CN`，当前不引入 i18n。

中文字体优先使用本机安装的思源黑体，随后回退到系统字体：

```css
"Source Han Sans SC", "Source Han Sans CN", "思源黑体",
"PingFang SC", "Microsoft YaHei", system-ui, sans-serif
```

全局样式和 Tailwind 的 `font-sans` 共用此字体栈。当前未内置字体文件；设备未安装思源黑体时使用系统字体，不依赖在线字体服务。

参考 [Next.js 官方文档](https://nextjs.org/docs)和 [create-next-app 参数说明](https://nextjs.org/docs/app/api-reference/cli/create-next-app)。

## 参与贡献

请阅读 [贡献指南](./CONTRIBUTING.md)。新功能开发前必须创建或复用对应 Issue；创建 PR 后，必须在 Issue 正文中用 `Related` 附上 PR 链接，并在 PR 描述中用 `Closes #编号` 引用已全部完成的 Issue，或用 `Related #编号` 引用仅关联或部分完成的 Issue。支持通过一个大 Issue 和多个子 Issue 拆分功能。

## 许可证

本项目采用 [MIT 许可证](./LICENSE)。
