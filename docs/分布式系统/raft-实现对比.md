# 四种语言 Raft 实现对比：etcd、MongoDB 与 Redis 的共识差异

> 素材来自源码走读笔记（etcd-io/raft · MongoDB pv1 · Redis Sentinel/Cluster）。标题写“四种语言”，正文以 Go / C++ / C 三套生产实现为主；Java 侧（Ratis / SOFAJRaft）放在最后作为选型参考。

## 1. 背景

在之前的项目里，多副本的数据一致性问题无法回避。团队内部经常能听到几种判断：

- “etcd 用的是 Raft，所以安全。”
- “MongoDB 副本集也是 Raft，主备自动切换。”
- “Redis Sentinel 和 Cluster 的选举机制跟 Raft 差不多。”

这些说法如果被当成同一种保证，选型或排查故障时就容易在写入丢失窗口、提交语义、成员变更这些地方出问题。我需要的不是重复论文内容，而是搞清楚：**它们各自到底是不是 Raft？具体差在哪？在默认配置下，客户端能依赖什么？**

为了回答这些问题，我对三份实现做了一次并排走读：`etcd-io/raft`（Go）、MongoDB `pv1`（C++）、Redis `sentinel.c` / `cluster_legacy.c`（C）。下面是对比记录。

## 2. 整体定性：是不是 Raft

很多系统提到“Raft”，其实只实现了选举和任期，而没有覆盖 Raft 安全性的两个核心约束：**日志的前缀匹配**，以及 **commit index 必须由当前任期的条目在多数派复制后才能推进**。这两点才是“已确认写入不丢失”的根基。

先把三者的基本情况交代清楚：

- **etcd-io/raft**：完整的 Raft 实现，纯状态机库，不负责 I/O，通过 `Ready` 结构体把需要处理的工作交给上层。
- **MongoDB pv1**：借鉴了任期和多数派选举，但复制机制是另一套基于 oplog 的拉取模型。我称之为“类 Raft”。
- **Redis Sentinel/Cluster**：Sentinel 的选举受 Raft 启发，但它没有日志，只是选出一个协调者；Cluster 使用 gossip 加 epoch 投票。两者都不是 Raft，官方也未曾声称它们是。

这个定性是后续讨论的基础。

## 3. 领导者选举

三者的选举都想实现同一个目标：在网络分区或节点故障时尽快选出新主，同时尽量避免多主。

Raft 协议里不存在“某个节点被判定为追随者无效”这样的外部裁决机制，节点的状态完全由自己根据收到的消息和超时事件来转变。
被隔离的节点只是暂时与集群失去联系，它本身依然是集群配置中的一员（除非显式通过配置变更移除它）。在采用标准 Raft 协议的分布式集群中，无论是领导者真的故障，还是追随者因网络隔离收不到心跳，选举超时都是触发领导者更替的必经步骤：真正的领导者失联会被淘汰，集群通过新一轮选举产生新领导者；而仅仅是某个追随者自己网络不通，这样的节点会因收不到心跳而超时，进而递增自己的 Term 并发起选举；等到网络恢复后，它携带的更高 Term 会迫使现任合法领导者下台，集群造成一次不必要的换届导致的服务中断。
如果一个节点已经通过配置变更被移出集群，那它既收不到领导者的心跳，其他节点也不会再向它复制日志。

etcd 的实现遵循 Raft 协议，并引入 PreVote 来避免上述 Term 无谓增长的问题。每个节点会将 Term 和投票记录持久化到 HardState，但在转为候选人并递增 Term 之前，会先发起 PreVote（MsgPreVote）。节点必须先探明自己能否获得多数派的预投票，只有通过这一轮“预投票”后，才真正增加 Term 并进入候选人状态。这样就避免了因网络隔离造成的 Term 无谓增长和集群振荡——被隔离的节点因无法联系多数派而根本不敢涨 Term，网络恢复时便能无声息地回归。

投票判定的核心逻辑如下（raft.go），其中 isUpToDate 检查候选人的日志是否足够新（最后一条日志的 term 和 index 均不落后于本地）：

```go
case pb.MsgVote, pb.MsgPreVote:
    // We can vote if this is a repeat of a vote we've already cast...
    canVote := r.Vote == m.GetFrom() ||
        // ...we haven't voted and we don't think there's a leader yet in this term...
        (r.Vote == None && r.lead == None) ||
        // ...or this is a PreVote for a future term...
        (m.GetType() == pb.MsgPreVote && m.GetTerm() > r.Term)
    // ...and we believe the candidate is up to date.
    lastID := r.raftLog.lastEntryID()
    candLastID := entryID{term: m.GetLogTerm(), index: m.GetIndex()}
    if canVote && r.raftLog.isUpToDate(candLastID) {
        // 授予投票...
    }
```

`isUpToDate` 就是日志新鲜度检查：candidate 的 `(term, index)` 必须不落后于本地日志的最后一条。

`MongoDB` 的核心工作是存储数据、响应查询。但为了实现“多副本 + 自动切主”，必须保证一条不变式：任意时刻最多只有一个节点能接受写入，防止两个节点同时认为自己是主而各自接收写入，导致数据不一致。在节点可能故障、网络可能自然分区的前提下，这只能依靠共识协议来保证。即便是“数据库”代码库，也需要一整套选举逻辑——`repl/` 目录正是集群模式下**数据正确性**的地基。  
具体实现位于 `mongo/db/repl/topology_coordinator.cpp`，负责 `mongod` 复制集的选举协调；另一个相关模块是分片，位于 `db/s` 下，详情参见 [raft-横向扩展对比](待写)。

`MongoDB` 的设计思路与 etcd 类似：etcd 使用 `PreVote` 预投票机制，`MongoDB` 则通过dry-run机制在调度层发起选举。且实现位置不同——etcd 的 `PreVote` 在追随者节点内部触发，而 `MongoDB` 的选举决策位于心跳触发的决策层，决策层实现时区分不同的心跳触发场景来达到类似的控制效果。
决策层共有四种触发场景：
- `kElectionTimeout`：心跳超时，一直未发现 primary；
- `kPriorityTakeover`：高优先级节点主动抢主；在基础raft增加的扩展
- `kCatchupTakeover`：节点追上进度后抢主；在基础raft增加的扩展
- `kStepUpRequestSkipDryRun`：手动执行 `rs.stepUp()` 时跳过 dry-run，直接发起真实选举。

只有最后一种场景会跳过 `dry-run` ；其余场景都会先发送一轮 `isADryRun=true` 的投票请求进行探路，探路成功后才发起真实选举——这就是代码中 `!args.isADryRun()` 判断的来源，用以区分 `dry-run` 预演和正式投票。  
这种 `dry-run` 机制,分两层，调度决策层在 `ElectionState::start()`（`replication_coordinator_impl_elect_v1.cpp`，协调/调度层），投票判定的实现是在`TopologyCoordinator::processReplSetRequestVotes`（决策层-集群拓扑协调器中），是一次应用层的预演。与 `PreVote` 一样，它能有效减少因网络隔离导致的无谓任期增长、CPU 浪费和脑裂风险。

投票条件方面，要求候选者的 `lastWrittenOpTime` 不小于投票者自己的 `lastWrittenOpTime`，这与 Raft 的日志新鲜度检查意图一致。同时，`_lastVote` 信息会持久化，确保每个任期只投一票。相关逻辑在 `topology_coordinator.cpp` 的 `processReplSetRequestVotes` 中：

```cpp
} else if (args.getLastWrittenOpTime() < getMyLastWrittenOpTime()) {
    response->setVoteGranted(false);
    response->setReason(fmt::format(
        "candidate's data is staler than mine. candidate's last written OpTime: {}, "
        "my last written OpTime: {}", ...));
} else if (!args.isADryRun() && _lastVote.getTerm() >= args.getTerm()) {
    response->setVoteGranted(false);
    response->setReason(fmt::format(
        "already voted for another candidate ({}) this term ({})", ...));
} else {
    ...
    if (!args.isADryRun()) {
        _lastVote.setTerm(args.getTerm());
        _lastVote.setCandidateIndex(args.getCandidateIndex());
    }
    response->setVoteGranted(true);
}
```

可以看到 `isADryRun()` 贯穿整个判定：dry-run 阶段不落盘 `lastVote`，只有真实选举才会持久化投票记录。

Redis Sentinel 本身没有预投票机制，而是通过随机选举超时来降低多个 Sentinel 同时成为候选人的概率，以此实现更公平的竞选。更关键的区别在于，Sentinel 选举 Sentinel leader 时不做日志新鲜度检查。它只关注谁先拿到多数票——哪个 Sentinel 获得了多数 Sentinel 的投票，谁就成为本轮故障转移的负责人。这种设计在只用来执行管理决策（而非存储数据）的哨兵集群中是合理的。

而 Sentinel 在选择要提升的新主节点时，并不依赖投票时的日志版本比对。它会从在线的从节点中，综合优先级、复制偏移量等因素，主动挑选一个数据尽可能新的节点作为新主，因此不会出现选出一个与多数节点网络延迟过高或数据明显落后的节点。

Redis Cluster 的故障转移也有类似特点。从节点发起选举时，会比较 configEpoch，同样没有 Raft 那样严格的日志完整性限制（即不强制要求候选者的日志至少和投票者一样新）。理论上，如果仅看投票条件，似乎可能选出一个数据落后的从节点，从而导致已确认的写入在后续被覆盖。但实际执行中，Cluster 在选择新主时，会考察候选从节点的状态及其与故障主节点的数据延迟，天然会倾向于数据更新、更健康的节点；同时，Redis 的异步复制本身允许少量数据丢失，协议范围内允许出现“选出一个落后节点覆盖掉多数节点已确认数据”的脑裂式覆盖。见第六节。

> 图示源文件见 `diagrams/raft-overview.excalidraw`（四图合一，选举在左上角），PNG 导出后嵌入正文。

## 4. 日志 / 数据复制

复制机制是三套系统分歧最大的地方。

etcd 采用 **推送模型**。leader 向每个 follower 发送 `MsgApp` 消息，其中包含 `(prevLogIndex, prevLogTerm)`。follower 校验自己日志中对应位置的条目是否与这两个值完全一致，一致才追加，否则拒绝。leader 根据拒绝信息回退 `Next` 索引，逐步找到匹配点，然后覆盖冲突部分。凭借这种前缀一致性检查，leader 可以通过 `progress tracker` 精确跟踪每个 follower 的 `Match` 和 `Next`，进而干净地推导哪些条目已复制到多数派、可以提交。follower 侧的核心逻辑（`raft.go` `handleAppendEntries`）：

```go
func (r *raft) handleAppendEntries(m *pb.Message) {
    a := logSliceFromMsgApp(m)
    if a.prev.index < r.raftLog.committed {
        r.send(&pb.Message{To: m.From, Type: pb.MsgAppResp.Enum(), Index: new(r.raftLog.committed)})
        return
    }
    if mlastIndex, ok := r.raftLog.maybeAppend(a, m.GetCommit()); ok {
        r.send(&pb.Message{To: m.From, Type: pb.MsgAppResp.Enum(), Index: new(mlastIndex)})
        return
    }
    // 日志不匹配：返回一个 (index, term) 猜测点，供 leader 回退 Next
    hintIndex := min(m.GetIndex(), r.raftLog.lastIndex())
    hintIndex, hintTerm := r.raftLog.findConflictByTerm(hintIndex, m.GetLogTerm())
    r.send(&pb.Message{
        To: m.From, Type: pb.MsgAppResp.Enum(), Index: new(m.GetIndex()),
        Reject: new(true), RejectHint: new(hintIndex), LogTerm: new(hintTerm),
    })
}
```

`maybeAppend` 内部先用 `matchTerm` 校验 `prevLogIndex/prevLogTerm`，不匹配就整批拒绝，这正是前缀一致性的落地代码。

MongoDB 是 **拉取模型**。secondary 通过 `OplogFetcher` 以类似 tailable cursor 的方式从同步源拉取 oplog。这给予了选同步源的灵活性，但 leader 无法直接知道每个副本的确切复制进度，只能依靠心跳中上报的 `OpTime` 来间接推算多数派提交点。条目标识是 `(Timestamp, term)`，没有全局 Raft index；一致性检查保证批次内 OpTime 单调，但不做 `prevLogTerm` 那样的强前缀匹配。发生冲突时不会精确截断，而是走 rollback 再重新拉取的流程。`OplogFetcher` 的类注释直接点明了这套拉取语义（`oplog_fetcher.h`）：

```cpp
/**
 * The oplog fetcher, once started, reads operations from a remote oplog using a tailable,
 * awaitData, exhaust cursor.
 *
 * The initial `find` command is generated from the last fetched optime.
 * ...
 * When there is an error, it will create a new cursor by issuing a new `find` command to the
 * sync source.
 */
class OplogFetcher : public AbstractAsyncComponent<...> { ... };
```

Redis 是经典的 **异步主备复制**。master 将命令流写入 backlog 缓冲区，slave 通过 `PSYNC` 从某个偏移量续传。消息中只有 RESP 命令和 `master_repl_offset`，没有 term，也没有日志共识。如果断开过久，就只能全量重同步。`masterTryPartialResynchronization` 里能看到判定依据只是 `replid`（`replication.c`）：

```c
int masterTryPartialResynchronization(client *c, long long psync_offset) {
    char *master_replid = c->argv[1]->ptr;
    /* Is the replication ID of this master the same advertised by the wannabe
     * slave via PSYNC? If the replication ID changed this master has a
     * different replication history, and there is no way to continue. */
    if (strcasecmp(master_replid, server.replid) &&
        (strcasecmp(master_replid, server.replid2) ||
         psync_offset > server.second_replid_offset))
    {
        ...
        goto need_full_resync;
    }
    ...
}
```

没有 term、没有 `(prevIndex, prevTerm)`，能续传只取决于 `replid` 相不相同、`offset` 是否还在 backlog 范围内。

对比可以看出，etcd 的推送 + 前缀检查为提交安全性提供了坚实基础；MongoDB 的拉取更灵活，但多数派提交点的推导路径更长；Redis 则是流式复制，没有共识日志层面的冲突解决。

> 图示源文件见 `diagrams/raft-overview.excalidraw`（四图合一，复制模型在右上角），PNG 导出后嵌入正文。

## 5. 提交语义

这一部分在实际使用中差别最大，也最容易因为默认配置而产生误解。

etcd 的 `commit index` 严格等于多数派节点上 `match` 的最大索引值，并且需要满足“当前任期的至少一条条目已被多数派复制”这一附加规则。这保证了一条日志一旦显示为 committed，就一定已持久化在多数节点的存储中，任何未来的 leader 都会包含这条日志。“当前任期”规则就写在 `maybeCommit` 里（`log.go` + `tracker.go`）：

```go
func (l *raftLog) maybeCommit(at entryID) bool {
    // NB: term should never be 0 on a commit because the leader campaigned at
    // least at term 1. But if it is 0 for some reason, we don't consider this a
    // term match.
    if at.term != 0 && at.index > l.committed && l.matchTerm(at) {
        l.commitTo(at.index)
        return true
    }
    return false
}

// Committed returns the largest log index known to be committed based on what
// the voting members of the group have acknowledged.
func (p *ProgressTracker) Committed() uint64 {
    return uint64(p.Voters.CommittedIndex(matchAckIndexer(p.Progress)))
}
```

`at.term` 必须等于 leader 自己的当前 term，这就是“只提交当前任期条目”的硬约束。

MongoDB 要获得接近的持久性保证，需要显式配置 `w: "majority"` 并关注 journal 相关参数。因为它的默认 write concern 是 `w: 1`，只表示 primary 自己确认了写入。此时若 primary 故障，已 ACK 的操作可能在新主上不存在，出现 rollback 丢数据的情况。MongoDB 内部有 majority commit point 的计算逻辑，也会考虑 term 边界，但这些机制只有在 write concern 设为 majority 时才会对客户端生效。对应的 term 边界检查（`topology_coordinator.cpp` `advanceLastCommittedOpTimeAndWallTime`）：

```cpp
// This check is performed to ensure primaries do not commit an OpTime from a previous term.
if (_iAmPrimary() && committedOpTime.opTime < _firstOpTimeOfMyTerm) {
    return false;
}
if (!_selfConfig().isArbiter() &&
    getMyLastWrittenOpTime().getTerm() != committedOpTime.opTime.getTerm()) {
    if (fromSyncSource) {
        committedOpTime = std::min(committedOpTime, getMyLastWrittenOpTimeAndWallTime());
    } else {
        // Ignoring commit point with different term than my lastWritten...
        return false;
    }
}
```

这段代码和 etcd 的 `at.term != 0 && l.matchTerm(at)` 是同一个意图的两种写法，只是 MongoDB 把它做成了可选（依赖 `w: "majority"` 是否被启用）。

Redis 不存在集群范围的 commit index。Sentinel 源码注释明确说明，ODOWN 只是一个弱法定人数信号，不保证强一致。在分区和故障转移场景下，客户端有可能收到旧 master 的写入成功回复，之后被重定向到新 master，而已回复成功的写入可能就此丢失。原文注释（`sentinel.c`）：

```c
/* Is this instance down according to the configured quorum?
 *
 * Note that ODOWN is a weak quorum, it only means that enough Sentinels
 * reported in a given time range that the instance was not reachable.
 * However messages can be delayed so there are no strong guarantees about
 * N instances agreeing at the same time about the down state. */
void sentinelCheckObjectivelyDown(sentinelRedisInstance *master) { ... }
```

注释已经把结论写得很直白：ODOWN 是“弱法定人数”，不是共识意义上的 commit。

简单总结：**“有副本集”与“写入已提交”是两回事，两者之间隔着一层配置语义。**

> 图示源文件见 `diagrams/raft-overview.excalidraw`（四图合一，提交语义在左下角），PNG 导出后嵌入正文。

## 6. 成员变更与脑裂

成员变更是共识协议中最复杂的部分之一。etcd 支持联合共识和单步 ConfChange，变更请求本身作为一条 Raft 日志来复制，确保所有节点在同一配置上达成一致。在脑裂场景下，失去多数派的旧 leader 无法继续提交任何新条目。变更条目和普通日志走的是同一条路径（`raft.go` `stepLeader`）：

```go
if e.GetType() == pb.EntryConfChange {
    ccc := &pb.ConfChange{}
    if err := proto.Unmarshal(e.GetData(), ccc); err != nil {
        panic(err)
    }
    cc = ccc
} else if e.GetType() == pb.EntryConfChangeV2 {
    ccc := &pb.ConfChangeV2{}
    ...
}
if cc != nil {
    alreadyPending := r.pendingConfIndex > r.raftLog.applied
    alreadyJoint := len(r.trk.Config.Voters[1]) > 0
    ...
}
```

`EntryConfChange`/`EntryConfChangeV2` 只是普通 Entry 的一种类型标记，复制、提交、拒绝的规则和数据日志完全一样——这就是“变更请求本身作为一条 Raft 日志”的具体含义。

MongoDB 的 `rs.reconfig()` 通过心跳传播新配置，没有采用联合共识。在极端多数派同时变更的场景下，风险会相对更高，但它仍然依赖多数派选举和 stepdown 来防止脑裂。心跳发现更新配置后走的是 `_scheduleHeartbeatReconfig`，而不是一条需要多数派提交的日志（`replication_coordinator_impl_heartbeat.cpp`）：

```cpp
case HeartbeatResponseAction::Reconfig:
    invariant(responseStatus.getStatus());
    _scheduleHeartbeatReconfig(lock, responseStatus.getValue().getConfig());
    break;
case HeartbeatResponseAction::RetryReconfig:
    _scheduleHeartbeatReconfig(lock, rsc);
    break;
```

新配置按 `(version, term)` 排序，谁的更新就采用谁的——这是“心跳传播”而不是“日志共识”的直接证据。

Redis Cluster 有一个值得注意的函数 `clusterBumpConfigEpochWithoutConsensus()`，其注释直接标明“无共识”。旧 master 在分区期间仍可接受写入，直到被新配置 epoch 重定向为止。这期间的写入没有强一致性保障。函数注释原文（`cluster_legacy.c`）：

```c
/* Important note: this function violates the principle that config epochs
 * should be generated with consensus and should be unique across the cluster. */
int clusterBumpConfigEpochWithoutConsensus(void) {
    uint64_t maxEpoch = clusterGetMaxEpoch();
    if (myself->configEpoch == 0 || myself->configEpoch != maxEpoch) {
        server.cluster->currentEpoch++;
        myself->configEpoch = server.cluster->currentEpoch;
        ...
        return C_OK;
    }
    return C_ERR;
}
```

“violates the principle that config epochs should be generated with consensus”——这是 Redis 官方代码注释自己写的，不是我的引申。

脑裂防护力度从强到弱大致是：**etcd > 正确配置 majority write concern 的 MongoDB > Redis**。

> 图示源文件见 `diagrams/raft-overview.excalidraw`（四图合一，脑裂对比在右下角），PNG 导出后嵌入正文。

## 7. 选型参考

基于这次走读，我整理了下面的判断，便于日后选型时对照。

- **etcd（或同样严格的 Raft 库）**：适合元数据管理、配置中心等对强一致有硬性要求的场景。已确认的写入就是多数派持久化。
- **MongoDB**：如果数据不能丢失，务必在客户端统一设置 `w: "majority"`，并理解 rollback 行为。不应该把“开了副本集”直接等同于“安全多副本”。
- **Redis**：适合缓存、会话等可重建数据。不要把它当强一致账本使用，故障转移窗口内的写入丢失是协议允许的。
- **Java 自研共识组件**：建议优先对照 Apache Ratis 或 SOFAJRaft，它们的模型接近 etcd 的推送 + progress tracker，而不是照着 Redis Cluster 的 epoch 投票去推导 Raft。

判断一个系统是不是真正实现了 Raft，最核心的检查点仍然是两个：是否强制做 `(prevIndex, prevTerm)` 前缀匹配；commit 推进是否要求当前任期的条目先被多数派复制。

MongoDB 和 Redis 不完全满足这两条。这并不是说它们“做得不好”，而是它们本来就不承诺和 Raft 一样的安全边界。协议形状不同，能给的承诺自然不同。

## 8. 后续怎么做

如果再次面对类似选型，我会先完成下面几件事：

1. 为每个候选组件整理一张“写入安全窗口”表，明确：默认 ACK 的含义、分区时旧主是否能写入、故障转移后未提交数据是覆盖还是 rollback。
2. 评审中不再接受“有选举”等同于“有 Raft”的说法，直接要求对方说明是否存在 `(prevIndex, prevTerm)` 检查、集群级 commit index、成员变更是否走共识日志。
3. 如果选 MongoDB 且要求不丢写，把 `majority` 作为基准配置，并通过故障演练验证 rollback 场景。
4. 若未来涉及嵌入式或自研共识库，以 etcd-io/raft 的 `Ready` + progress tracker 为工程参照。Java 栈直接深入 Ratis/JRaft。
5. 另找时间写一篇 Java 侧的对比，详细拆解 Ratis / SOFAJRaft 在 API 与线程模型上与 etcd 的差异。这篇文章只把结论放在这里。

---

## 9. 附录：更多源码片段

正文里的引用已经够支撑每一节的结论，下面这些是走读时顺带记的补充片段，感兴趣可以展开看，不影响正文的完整性。

### 9.1 选举补充：Redis Sentinel 的随机延迟

第 3 节提到 Sentinel 用随机延迟降低多候选人冲突，具体是这一句（`sentinel.c` `sentinelStartFailover`）：

```c
void sentinelStartFailover(sentinelRedisInstance *master) {
    serverAssert(master->flags & SRI_MASTER);
    master->failover_state = SENTINEL_FAILOVER_STATE_WAIT_START;
    master->flags |= SRI_FAILOVER_IN_PROGRESS;
    master->failover_epoch = ++sentinel.current_epoch;
    sentinelEvent(LL_WARNING,"+new-epoch",master,"%llu",
        (unsigned long long) sentinel.current_epoch);
    sentinelEvent(LL_WARNING,"+try-failover",master,"%@");
    master->failover_start_time = mstime()+rand()%SENTINEL_MAX_DESYNC;
    master->failover_state_change_time = mstime();
}
```

`rand()%SENTINEL_MAX_DESYNC` 直接加在启动时间上，纯粹是为了错开多个 Sentinel 同时发起 failover，跟日志/数据没有任何关系。

### 9.2 复制补充：etcd 的 Match/Next 与冲突回退

第 4 节提到 leader 靠 `progress tracker` 跟踪每个 follower 的 `Match`/`Next`（`tracker/progress.go`）：

```go
type Progress struct {
    // Match is the index up to which the follower's log is known to match the leader's.
    Match uint64
    // Next is the log index of the next entry to send to this follower. All
    // entries with indices in (Match, Next) interval are already in flight.
    //
    // Invariant: 0 <= Match < Next.
    Next uint64
    // ...
}

// MaybeUpdate is called when an MsgAppResp arrives from the follower.
func (pr *Progress) MaybeUpdate(n uint64) bool {
    if n <= pr.Match {
        return false
    }
    pr.Match = n
    pr.Next = max(pr.Next, n+1)
    pr.MsgAppFlowPaused = false
    return true
}
```

而 `findConflictByTerm` 的文档注释解释了“回退 Next 找匹配点”具体在猜什么（`log.go`）：

```go
// findConflictByTerm returns a best guess on where this log ends matching
// another log, given that the only information known about the other log is the
// (index, term) of its single entry.
//
// Specifically, the first returned value is the max guessIndex <= index, such
// that term(guessIndex) <= term or term(guessIndex) is not known (because this
// index is compacted or not yet stored).
func (l *raftLog) findConflictByTerm(index uint64, term uint64) (uint64, uint64) { ... }
```

### 9.3 成员变更补充：MongoDB 心跳 reconfig 的并发保护

第 6 节提到 MongoDB 靠心跳传播配置、没有联合共识；对应地，它也没有 Raft 那种“配置变更本身要走多数派提交”的保护，只能用一个状态机粗粒度地防止并发 reconfig（`replication_coordinator_impl_heartbeat.cpp` `_scheduleHeartbeatReconfig`）：

```cpp
switch (_rsConfigState) {
    case kConfigUninitialized:
    case kConfigSteady:
        LOGV2_FOR_HEARTBEATS(4615622, 1, "Received new config via heartbeat", ...);
        break;
    case kConfigInitiating:
    case kConfigReconfiguring:
    case kConfigHBReconfiguring:
        LOGV2_FOR_HEARTBEATS(4615623, 1,
            "Ignoring new configuration because we are already in the midst of a configuration "
            "process", ...);
        return;
    ...
}
```

这是一个本地状态标志位，不是共识协议——和 etcd 的 ConfChange 走日志复制相比，安全边界明显更弱。

---

> 这次走读没有做性能压测，所有结论都基于源码结构和协议路径的梳理。对我来说，这已经足够回答最初的问题：**在不同多副本方案面前，我可以相信什么。**

**相关阅读（站内规划）**
- 从 Raft 论文到生产实现：工程师阅读路径（待写）
- 消息中间件选型：Redpanda 能替代什么（待写）