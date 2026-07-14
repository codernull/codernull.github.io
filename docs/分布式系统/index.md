# 分布式系统

共识、复制、消息中间件。

## 文章

| 文章 | 一句话结论 |
|---|---|
| [四种语言 Raft 实现横向对比](./raft-实现对比) | etcd 是严格 Raft；MongoDB pv1 是类 Raft；Redis Sentinel/Cluster 不是 Raft |

## 计划中

- 消息中间件选型：Redpanda 能替代什么，不能替代什么
  > 从存储模型、延迟特征和运维成本几个角度，分析 Redpanda 相比 Kafka 的优势与边界。
- 从 Raft 论文到生产实现：一份工程师视角的阅读路径
  > 梳理从原始论文到 etcd-io/raft 等生产级代码的阅读线索，帮助把理论映射到工程实现上。
