---
layout: home

hero:
  name: false
  text: false
  tagline: 分布式系统 · 低延迟设计 · Go 工程 · 工程方法
  actions:
    - theme: brand
      text: 阅读代表作：Raft 实现对比
      link: /分布式系统/raft-实现对比
    - theme: alt
      text: 浏览全部分类
      link: /分布式系统/

features:
  - title: 分布式系统
    details:  共识、复制、消息中间件-共识协议源码级对比与生产取舍
    link: /分布式系统/
  - title: 低延迟系统
    details: 行情推送、抖动与唤醒路径 — 从测量数据到优化决策
    link: /低延迟系统/
  - title: Go 实战
    details: WebSocket、WSL + Neovim — 真实工程约束下的实践
    link: /Go实战/
  - title: 工程方法
    details: 规划、复盘、看板 — 把做事方法沉淀为可复现的结论
    link: /工程方法/
---

## 关于作者

<!-- 以下条目由 AI 基于站内素材起草的脱敏示例，请自行核实后替换为真实内容；数字建议只给数量级或标注「估算」。 -->

专注分布式共识与低延迟系统设计，长期解决多副本一致性与高性能推送相关的工程问题。

- 主导某交易系统多副本架构落地，用 Raft 共识组 + 异步流水线把故障切换从秒级压到亚秒级，满足强一致要求。
- 负责某订阅中心存储层换代，用日志结构存储替代关系型方案，写入吞吐提升约一个数量级，P99 延迟进入毫秒级。
- 在行情推送热路径上引入 Aeron IPC，把唤醒与抖动控制在微秒级，并沉淀出可复用的选型判断。
- 做过 Kafka / Redpanda / Aeron 的消息中间件横向选型，从协议与源码层面给出「能替代什么、不能替代什么」的结论。

联系方式：[GitHub](https://github.com/codernull) · 邮箱 `your-email@example.com`
