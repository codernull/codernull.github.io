# codernull.github.io

[VitePress](https://vitepress.dev/) 技术文档站：**结论型深度文章**，面向搜索特定技术关键词的开发者。

站点：<https://codernull.github.io>

## 内容原则

每篇文章固定四段：

1. 我面对的真实问题
2. 我调研了什么、踩了什么坑
3. 最终结论和数据（估算会标明）
4. 如果再来一次我会怎么做

不做入门教程。

## 站点结构

```text
docs/
├── 分布式系统/     # Raft、共识、消息中间件选型
├── 低延迟系统/     # 行情、抖动、唤醒路径
├── Go实战/         # WebSocket、WSL + Neovim
└── 工程方法/       # 规划可见化、复盘、看板
```

## 本地开发

需要 Node.js 18+。

```bash
npm install
npm run docs:dev
```

构建：

```bash
npm run docs:build
npm run docs:preview
```

## 部署

推送到 `main` 后，GitHub Actions（`.github/workflows/deploy.yml`）构建并发布到 GitHub Pages。

仓库 Settings → Pages → Source 选 **GitHub Actions**。

## 已发布

| 文章 | 路径 |
|---|---|
| 四种语言 Raft 实现横向对比 | [`docs/分布式系统/raft-实现对比.md`](docs/分布式系统/raft-实现对比.md) |
| WSL + Neovim 配置全记录（可复现，不是插件安利） | [`docs/Go实战/wsl-neovim-配置全记录.md`](docs/Go实战/wsl-neovim-配置全记录.md) |
| 撮合引擎核心：订单簿结构、撮合语义与并发读取 | [`docs/低延迟系统/撮合引擎核心.md`](docs/低延迟系统/撮合引擎核心.md) |

素材来自 Dendron：`consensus.raft.comparison.md`（etcd / MongoDB / Redis 源码级对比）。
素材来自 WSL Ubuntu 用户配置复查（用户名、目录已脱敏）。
素材来自本仓库 `matching-engine`（Go 撮合引擎，源码与压测均可复现）。
