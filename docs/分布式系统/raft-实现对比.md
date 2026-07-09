# 四种语言 Raft 实现横向对比

> 素材来自源码走读笔记（etcd-io/raft · MongoDB pv1 · Redis Sentinel/Cluster）。标题写「四种语言」，正文以 **Go / C++ / C** 三套生产实现为主；Java 侧（Ratis / SOFAJRaft）放在「如果再来一次」里作为落地参照。

## 1. 我面对的真实问题

做低延迟行情与订阅中心时，绕不开「多副本怎么保证不丢、不脑裂」。团队里常听到三种说法：

- 「etcd 用的是 Raft，所以安全」
- 「MongoDB 副本集也是 Raft」
- 「Redis 有 Sentinel / Cluster，选举也像 Raft」

选型或排障时，如果把这三者当成同一种共识，会在 **写入丢失窗口**、**提交语义**、**成员变更** 上踩坑。我需要的不是论文复述，而是：**它们各自到底是不是 Raft、差在哪、默认配置下客户端能信什么**。

于是对照三份实现做了并排走读：`etcd-io/raft`（Go）、MongoDB `pv1`（C++）、Redis `sentinel.c` / `cluster_legacy.c`（C）。

## 2. 我调研了什么、踩了什么坑

### 2.1 总览：先分清「是不是 Raft」

| 维度 | **etcd-io/raft（Go）** | **MongoDB pv1（C++）** | **Redis Sentinel/Cluster（C）** |
|------|----------------------|----------------------|-------------------------------|
| **是否真正 Raft** | ✅ 完整 Raft（论文严格实现） | 🟡 类 Raft（借鉴任期与多数派选举，复制机制不同） | ❌ 非 Raft（Sentinel：受 Raft 启发的协调者选举；Cluster：gossip + epoch 投票） |
| **核心抽象** | 纯状态机库，不做 I/O，经 `Ready` 交付工作包 | `ReplicationCoordinatorImpl` + `TopologyCoordinator` 驱动完整服务器逻辑 | `sentinel.c` 单体状态机 + `cluster_legacy.c` gossip 循环 |

**踩坑点**：口头上的「我们有 Raft」往往只覆盖了「有选举 / 有任期」，没有覆盖 **日志匹配性质** 和 **commit index 规则**——后两者才是「已确认写入不会丢」的根基。

### 2.2 领导者选举

| 维度 | **etcd** | **MongoDB** | **Redis** |
|------|----------|------------|----------|
| **任期 / 纪元** | `Term`，持久化在 `HardState` | `OpTime.term`；`TopologyCoordinator::_term`；`lastVote` 落盘 | Sentinel：`current_epoch`；Cluster：`currentEpoch` + `configEpoch` |
| **预选举** | **PreVote**（`MsgPreVote`） | **dry-run election**（应用层预演） | 无；Sentinel 用随机延迟降冲突 |
| **投票条件（日志新鲜度）** | `isUpToDate`：先比 last term，再比 last index | `lastWrittenOpTime >=` 投票者侧 | Sentinel：无日志新鲜度；Cluster：偏 configEpoch |
| **每任期一票** | ✅ `r.Vote` + `HardState` | ✅ `LastVote` 落盘后才发真实投票 | ✅ epoch 守卫 |

**结论片段**：etcd 的 PreVote 与 MongoDB 的 dry-run 目标相同——避免无谓 term 递增导致振荡；路径不同：一个是协议消息，一个是应用层预演。Redis **不做日志新鲜度限制**，只选「谁拿到多数票」，这是安全性弱于 Raft 的原因之一。

### 2.3 日志 / 数据复制：推送 vs 拉取

| 维度 | **etcd** | **MongoDB** | **Redis** |
|------|----------|------------|----------|
| **推 / 拉** | **推送**：leader 发 `MsgApp` | **拉取**：secondary `OplogFetcher` + tailable cursor | **推送**：`replicationFeedSlaves` 异步 backlog |
| **条目标识** | `(index, term)` protobuf Entry | `OpTime = (Timestamp, term)`，无全局 Raft index | RESP 命令流 + `master_repl_offset`，无 term |
| **一致性检查** | `(prevIndex, prevTerm)` 必须匹配 | 批次内 OpTime 单调；无 prevLogTerm 式强前缀检查 | 无复制日志共识；PSYNC 断点续传 |
| **冲突处理** | 截断覆盖 + Next/Match 探测 | **rollback** 后重新拉取 | full resync / PSYNC，无按条目冲突解决 |

**踩坑点**：推送模式下 leader 能精确维护 `Match/Next`（`tracker/progress.go`），commit 推导干净；拉取模式下 secondary 自选 sync source，leader 只能靠心跳里的 OpTime **间接**推多数派提交点——更灵活，安全性证明更绕。Redis 是流式主备，不是共识日志。

### 2.4 提交 / 持久化：差距最大的一维

| 维度 | **etcd** | **MongoDB** | **Redis** |
|------|----------|------------|----------|
| **提交语义** | `commit index` = 多数派 match 的最大 index + **当前任期提交规则** | majority commit point（投票节点 OpTime 的第 majority 小值）+ term 边界 | 无集群范围 commit index |
| **多数派写入** | ✅ 多数派持久化后才 committed | 🟡 需 `w:"majority"`（及 journal 相关配置） | ❌ 默认异步；ACK 后 replica 可能未收到 |
| **写入丢失风险** | 已提交不可覆盖 | 默认 `w:1` 故障转移可能 rollback | 高：分区 + failover 可丢已 ACK 写入 |

**踩坑点（最常见）**：把 MongoDB 副本集默认当成「安全多副本」。默认 `w:1` 只等 primary 确认；只有 `w: "majority"`（并配合 journal / `writeConcernMajorityShouldJournal`）才接近 Raft commit 的持久性。Redis Sentinel 源码注释也写明 ODOWN 是 **弱法定人数信号**，不保证强一致。

### 2.5 成员变更与脑裂

| | **etcd** | **MongoDB** | **Redis** |
|---|----------|------------|----------|
| **成员变更** | 联合共识 / Simple ConfChange，走 Raft 日志 | `rs.reconfig()`，心跳传播，无 joint consensus | Cluster 可 `clusterBumpConfigEpochWithoutConsensus()`——注释标明「无共识」 |
| **防脑裂** | quorum + term；无多数派无法提交 | 多数派选举 + stepdown | 旧 master 可继续写直到被重定向；窗口内写入可能丢 |

脑裂防护力度大致是：**etcd > MongoDB（majority WC）> Redis**。

## 3. 最终结论和数据

> 下列「数据」主要是**结构结论与源码锚点**，不是压测数字。延迟/吞吐需按业务单独测。

### 3.1 一句话结论

1. **etcd-io/raft**：真正 Raft——日志匹配性质 + 当前任期提交规则成立；「已确认 ≈ 已提交 ≈ 多数派持久化」。
2. **MongoDB pv1**：类 Raft——选举与 term 像 Raft，复制是 **oplog 拉取 + rollback**；安全性取决于客户端 **writeConcern**，不是「开了副本集就安全」。
3. **Redis Sentinel/Cluster**：不是 Raft——选举协调 + 异步复制；故障转移场景下 **已 ACK 写入可能丢失**。
4. **Java 生态**（未在本文逐行走读）：Apache Ratis、SOFAJRaft 走 **推送 + progress tracker** 路线，工程模板更接近 etcd，而不是 MongoDB 拉取模型。

### 3.2 「真正 Raft」只认两件事

- AppendEntries 的 **日志匹配性质**（前缀一致）
- commit 推进遵守 **当前 term 条目先复制到多数派**，旧 term 只能间接提交

MongoDB / Redis 都不完整满足这两点，因此在极端分区下有不同程度的数据风险——这不是「实现质量差」，而是 **协议形状不同**。

### 3.3 选型速查（估算级判断）

| 你的约束 | 更合理的默认 |
|---|---|
| 元数据 / 配置中心，强一致优先 | etcd（或同等严格 Raft 库） |
| 文档库，可接受显式 majority WC | MongoDB + `w:"majority"` |
| 缓存 / 会话 / 可丢可重建 | Redis；别当强一致账本 |
| 自研 Java 共识组件 | 优先对照 Ratis / JRaft（推送模型），不要照搬 Redis 选举叙事 |

## 4. 如果再来一次我会怎么做

1. **先画「写入安全窗口」表**：对每个候选组件写清——默认 ACK 是否等于多数派持久化；分区时旧主能否继续写；failover 后未提交数据走覆盖还是 rollback。
2. **禁止用「有选举」当「有 Raft」**：评审里强制对照：有没有 `(prevIndex, prevTerm)`、有没有 commit index、成员变更是否走共识日志。
3. **MongoDB 默认配置当红线**：业务若不能丢写，配置与客户端 SDK 统一 `majority`，并在故障演练里验证 rollback 行为。
4. **自研或嵌入式共识**：以 etcd-io/raft 的 `Ready` + progress tracker 为模板；Java 直接读 Ratis/JRaft，而不是从 Redis Cluster 的 epoch 投票「反向发明 Raft」。
5. **补第四种语言的对照篇**：单独写一篇 Ratis / SOFAJRaft 与 etcd 的 API/线程模型差异（本次只落到选型结论，未展开源码表）。

---

**相关阅读（站内规划）**

- 从 Raft 论文到生产实现：工程师阅读路径（待写）
- 消息中间件选型：Redpanda 能替代什么（待写）
