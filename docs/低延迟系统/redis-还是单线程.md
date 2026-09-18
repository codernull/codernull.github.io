# Redis 还是单线程：6.0/8.x 改了外围，热路径不该指望它

## 0. 核心结论

Redis 的"单线程"指命令执行始终在主线程串行：6.0 IO 线程只搬读写不搬执行，8.x 重写实现后不变量没动。热路径延迟瓶颈在主线程时，升版本救不了；Redis 的正确层位是字典点查，事件分发不归它管。

## 1. 背景与动机

做低延迟热路径时，团队里关于 Redis 的判断经常是这几种：

- "升 6.0，IO 多线程，吞吐翻倍。"
- "`io-threads` 打开就行，配置里改个数。"
- "8.x 延迟更低，直接上。"
- "Pub/Sub 扛不住？上集群。"

这些说法共享一个隐含前提：**Redis 的"快"是可以靠版本和配置升级的**。但它们从来没回答一个问题——6.0 的 IO 线程到底搬走了什么？8.x 重写实现后，那条最关键的线有没有动？

我需要的不是"Redis 很快"的重复，而是搞清楚：**在低延迟热路径上，Redis 的延迟上限由什么决定，升级到底能不能移动它**。为了回答这个问题，我 pinned 了两个 tag（`6.0.0`、`8.8.0`）逐函数走读，把"哪个线程在扛"这件事画成了执行边界图。

本文的结论只在单机热路径层位成立：不覆盖集群多活、不覆盖大 Value 场景、不覆盖异步复制语义。

## 2. 单线程的真相：一个进程里其实有四种线程

"Redis 是单线程" 它特指**命令执行**，不是进程只有一条线程——从 2.4 起 Redis 就有后台线程（`bio.c`），4.0 加了 lazyfree 异步释放，6.0 又多了 IO 线程。

`bio.c`（tag `6.0.0`）启动时按任务类型各建一条线程：

```c
/* bio.c bioInit, tag 6.0.0 */
for (j = 0; j < BIO_NUM_OPS; j++) {
    void *arg = (void*)(unsigned long) j;
    if (pthread_create(&thread,&attr,bioProcessBackgroundJobs,arg) != 0) {
        serverLog(LL_WARNING,"Fatal: Can't initialize Background Jobs.");
        exit(1);
    }
    bio_threads[j] = thread;
}
```

每条后台线程只处理一种任务（`bio.c` `bioProcessBackgroundJobs`，tag `6.0.0`）：

```c
switch (type) {
case BIO_CLOSE_FILE:
    redis_set_thread_title("bio_close_file");
    break;
case BIO_AOF_FSYNC:
    redis_set_thread_title("bio_aof_fsync");
    break;
case BIO_LAZY_FREE:
    redis_set_thread_title("bio_lazy_free");
    break;
}
```

三类任务：关闭文件、AOF fsync、lazyfree 内存释放。它们的共同点是**慢 I/O 或慢释放**。为什么必须是这三类？把 fsync 放回主线程推演一遍就清楚了：一次 fsync 可能耗时几十毫秒，如果主线程等它落盘，这几十毫秒里所有客户端的命令全部排队——一条慢 fsync 就能把整个实例卡成假死。BIO 线程把它们搬走，主线程只负责"把任务入队"这一个动作。

但要注意边界：**BIO 里没有任何数据操作**。命令的读写不经过 BIO 线程，它们只处理"文件/内存的收尾"。

所以一个 6.0 进程的线程构成是：主线程（事件循环 + 命令执行）、`bio_close_file`、`bio_aof_fsync`、`bio_lazy_free`。加上 IO 线程（下一节），默认 6 条线程——但命令执行始终只有主线程那一条。

这四种线程的分工里，前三种不碰数据，真正值得细看的是 6.0 加的第四种。它到底搬了什么，是这篇文章第一个要核对的点。

> 图示源文件见 `diagrams/redis-single-thread.excalidraw`（四种线程与读写/执行边界，PNG 导出后嵌入正文）。

## 3. 6.0 IO 多线程：搬了读写，没搬执行

6.0 引入 IO 线程的动机是网卡变快后网络读写成为新瓶颈。关键在"搬了什么"：**读写搬走了，执行没搬**。

IO 线程的工作循环（`networking.c` `IOThreadMain`，tag `6.0.0`）只做两件事：

```c
while((ln = listNext(&li))) {
    client *c = listNodeValue(ln);
    if (io_threads_op == IO_THREADS_OP_WRITE) {
        writeToClient(c,0);
    } else if (io_threads_op == IO_THREADS_OP_READ) {
        readQueryFromClient(c->conn);
    } else {
        serverPanic("io_threads_op value is unknown");
    }
}
```

`writeToClient` 把回复写回客户端 socket，`readQueryFromClient` 把请求读进缓冲区。**两者都不执行命令**。命令执行发生在哪？把一次请求的完整链路走一遍，边界就清楚了：

```text
1. IO 线程：readQueryFromClient  把请求读进客户端的输入缓冲
2. 主线程：processInputBuffer    从缓冲解析命令
3. 主线程：processCommand        执行——ACL 检查、命令表分发、读数据、把回复写进输出缓冲
4. 事件循环：writeToClient       把输出缓冲写回 socket（可交给 IO 线程）
```

第 3 步是唯一的"执行"，从 2.4 到 8.8 都没离开过主线程。IO 线程能把第 1、4 步搬走，搬不走第 3 步——执行回主线程这件事，源码里写在 `handleClientsWithPendingReadsUsingThreads` 的收尾：IO 线程把读完成之后，主线程重新跑一遍客户端列表（`networking.c`，tag `6.0.0`）：

```c
/* Run the list of clients again to process the new buffers. */
while(listLength(server.clients_pending_read)) {
    ln = listFirst(server.clients_pending_read);
    client *c = listNodeValue(ln);
    c->flags &= ~CLIENT_PENDING_READ;
    listDelNode(server.clients_pending_read,ln);

    if (c->flags & CLIENT_PENDING_COMMAND) {
        c->flags &= ~CLIENT_PENDING_COMMAND;
        if (processCommandAndResetClient(c) == C_ERR) {
            continue;
        }
    }
    processInputBuffer(c);
}
```

注释已经把分工写明白了：IO 线程把数据放进缓冲区，主线程"再跑一遍"去 `processCommandAndResetClient` + `processInputBuffer`。`processCommandAndResetClient` 内部就是 `processCommand`（命令分发、ACL 检查、执行、写回复队列），全部在主线程。

还有两个配置细节，同样是边界的证据：

- `postponeClientRead`（`networking.c`，tag `6.0.0`）只有在 `server.io_threads_do_reads` 为真时才把读请求推迟给 IO 线程——而 `io-threads-do-reads` 默认是 `no`，即**默认连读都不搬，只有写回走 IO 线程**。为什么官方默认不搬读？因为搬读之后主线程要等 IO 线程完成才能执行，多一次同步；小包点查场景里搬读往往得不偿失。
- `stopThreadedIOIfNeeded`（`networking.c`，tag `6.0.0`）在 `pending < io_threads_num*2` 时停掉 IO 线程回到同步写——**低负载自动回退单线程**。这说明 IO 线程的定位是"高峰辅助"，不是常态执行通道。

这三个证据（IO 线程只读写、执行回主线程、低负载回退）合起来就是：IO 线程是命令执行的**辅助通道**，不是替换。

### 版本线核对：8.x 函数搬了家，语义没变

8.8 里这套代码从 `networking.c` 搬到了独立的 `iothread.c`，而且实现重写了——每个 IO 线程有自己的事件循环，用 event notifier 和主线程通信（tag `8.8.0`）：

```c
/* iothread.c IOThreadMain, tag 8.8.0 */
void *IOThreadMain(void *ptr) {
    IOThread *t = ptr;
    ...
    aeSetBeforeSleepProc(t->el, IOThreadBeforeSleep);
    aeSetAfterSleepProc(t->el, IOThreadAfterSleep);
    aeMain(t->el);
    return NULL;
}
```

表面看"IO 线程都有自己的事件循环了"，像是多线程服务的模样。但注释把执行位置写死了（`iothread.c`，tag `8.8.0`）：

```c
/* IO-thread reads may enqueue one-by-one complete commands that are
 * executed in main thread without re-entering processInputBuffer(). */
```

"读进来、在主线程执行"——和 6.0 是同一句话。核对结论：**函数搬了家、实现重写了，命令执行仍在主线程这条不变量，跨 6.0→8.8 没动**。这是"升级能不能救延迟"的第一条证据。

## 4. 压测：瓶颈有没有搬家

这篇文章的高潮不是"开 4 个 IO 线程 GET 变快了多少"，而是验证一件事：**IO 线程到底把瓶颈搬走没有**。

实验设计上，每个场景对应源码里的一个具体假设：

| 实验 | 验证的假设 |
|---|---|
| 单客户端干净延迟 | 执行路径有没有被 IO 线程改变（§3 结论的正面验证） |
| c50 高并发 GET 小包 | IO 线程抬不抬执行吞吐（执行密集负载） |
| 4KB value GET | 写回成本最高的场景，IO 线程最该受益的地方 |
| PUBLISH 多进程发布 | 发布扇出执行在主线程，验证"发布路径挂主线程" |

实验在 WSL + localhost 上跑，Redis 按 tag `6.0.0` 现编（`MALLOC=libc`），三档配置：`io-threads 1`（即关闭）、`io-threads 4`、`io-threads 4 + io-threads-do-reads yes`。

### 4.1 单客户端干净延迟：三档没有差异

单客户端、服务端不饱和，测的是"执行路径本身的延迟"：

| 配置 | p50 | p99 | p999 |
|---|---|---|---|
| `io-threads 1` | 0.035 ms | 0.094 ms | 0.141 ms |
| `io-threads 4` | 0.038 ms | 0.096 ms | 0.159 ms |
| `io-threads 4` + do-reads | 0.035 ms | 0.086 ms | 0.136 ms |

三档几乎一致（35–38 us）。这直接对应源码：单请求的路径是"读 → 主线程执行 → 写回"，IO 线程不改变这条路径——它要么在帮忙搬读写，要么闲置，执行时间没动。

### 4.2 高并发吞吐：io-threads 4 没有让 GET 更快

`redis-benchmark -c 50 -n 500000`，GET 3B 小包（执行密集负载）：

| 配置 | rps | 尾部延迟 |
|---|---|---|
| `io-threads 1` | 170,823 | 99.93% ≤ 0.8 ms，100% ≤ 8 ms |
| `io-threads 4` | 127,975 | 99.93% ≤ 0.9 ms，**但 99.98% ≤ 338 ms、99.99% ≤ 662 ms** |
| `io-threads 4` + do-reads | 166,555 | 99.95% ≤ 0.8 ms，99.97% ≤ 9 ms |

在本机（WSL）环境下，`io-threads 4` 反而更慢，且出现严重长尾（600+ ms）。pipeline 场景同样没有帮助（`-P 16`，执行饱和）：`io1` 267 万 rps、`io4` 238 万、`io4`+do-reads 255 万——差异在抖动范围内，IO 线程没有抬升执行吞吐。

### 4.3 大 value / PUBLISH：IO 线程也没有帮上忙

4KB value 的 GET（写回成本高，IO 线程应受益的场景），4 进程并发实测：

| 配置 | 合计 rps | p50 | 主线程 CPU | IO 线程 CPU |
|---|---|---|---|---|
| `io-threads 1` | ~69 k | 0.52 ms | **68.5%** | — |
| `io-threads 4` | ~51 k | 0.65 ms | **46.0%** | 1.9–2.1% × 3 |
| `io-threads 4` + do-reads | ~51 k | 0.65 ms | **54.5%** | 2.3–2.5% × 3 |

PUBLISH（1 个订阅者 + 4 个发布进程，40 并发）：`io1` 合计 ~45 k rps、主线程 CPU **64.6%**；`io4` 合计 ~38 k rps、主线程 CPU **53.9%**（IO 线程各 6–7%）。

### 4.4 判断：瓶颈没有搬家

三条证据合起来，结论是同一个：

1. **主线程 CPU 始终是大头**（46–69%），IO 线程合计不到 10%——瓶颈还挂在执行线程上；
2. **单客户端延迟三档无差异**——执行路径没被 IO 线程改变；
3. **IO 线程没有在任何一档提升吞吐**，在 GET 小包和 PUBLISH 上反而更慢——调 `io-threads` 在本场景是安慰剂，不是解药。

`io4` 在本机反而更慢，大概率是 WSL 虚拟化网络栈 + 线程唤醒开销叠加的结果，但无论归因到哪，它都没有改变那条主线判断：**瓶颈在"命令执行"这条线上，不在读写**。

> **环境与限制（必须读）**：实验在 WSL + localhost 上跑，没有真实网卡排队、没有跨机延迟。这些数字只能回答"瓶颈有没有搬家"，**不能证明机房延迟**；`io-threads 4` 在本机反而更慢、长尾更差，也可能与 WSL 的虚拟化网络栈和线程唤醒开销有关，不能外推为生产结论。没有跨机、没有生产网卡，就写"没有测"——本文不会把本机能跑包装成"升 8.x 延迟更低"。全量数据见附录。

## 5. 8.x：改了外围，不变量没动

8.x 的变化很多，但都可以归进外围：模块内置（8.0 起 Search/JSON/TimeSeries 不再需要单独装模块）、新数据结构（vector set、8.8 的 Array）、性能改进（延迟降低、内存压缩）——**没有一条改变"命令执行单线程"这条不变量**。

为什么 8.x 在这里只占一小节？因为判断版本升级的锚点只有一个：那条不变量动没动。逐条核对：

1. **实现层面**：如第 3 节，IO 线程实现重写、代码搬家，执行仍回主线程。
2. **许可证层面**：Redis 8.x 的源码头不再是 BSD-3-Clause，改成 RSALv2 / SSPLv1 / AGPLv3 三选一（`iothread.c` 文件头，tag `8.8.0`）。这直接影响"要不要升 8.x"的评估，但和延迟无关。

所以"该不该上 8.x"的判断要分项：

- **可以评**：缓存点查的运维（8.x 内置模块省打包）、许可策略（RSALv2 是否可接受）、新数据结构的可用性（向量检索等具体需求）。
- **证据不够、动机不对**：为了延迟升 8.x。单机热路径的延迟上限由"命令执行单线程"决定，这一条 6.0 和 8.8 完全一样，升级不会搬走执行瓶颈。

## 6. 我的场景：Redis 停在字典点查层

（脱敏）在某新股申购系统里，Redis 的层位是三件事定下来的：

1. **额度冻结/扣减用 Redis 原子操作**（`INCRBY`）——这是字典点查：一个 key、一次操作、毫秒级完成，Redis 的主线程语义在这个量级没有任何压力。
2. **持仓用批量写本机快照 + 本地缓存预热**——Redis 不当地存储，快照落库是异步的；热点数据预热进服务本地缓存，读路径甚至不经过 Redis。
3. **事件流走 Kafka 异步削峰**——申购请求入队、消费者异步处理。订阅关系、消息风扇都不依赖 Redis Pub/Sub。

这三条反推回第 2、3 节的结论，根因是同一个：**Redis 的命令执行单线程在突发下顶不住**——大量并发写/PUBLISH 会把主线程打满，p50 稳定、p99 失控。所以凡是"高频事件分发"的活，都不放 Redis；凡是"单点状态查询/扣减"的活，Redis 是天然位置。

（第 4.3 节 PUBLISH 实测：40 并发发布下主线程 CPU 已 64.6%、p50 0.68 ms——发布路径确实挂在主线程上；若手头有"5 万订阅者"规模的线上数字，可再补进本节。）

## 7. 边界与选型判断

什么该放 Redis、什么不该放，边界其实很清晰：

| 该放（字典点查） | 不该放（事件/写入流） |
|---|---|
| 键值点查、计数器、限流（INCR/INCRBY） | 高频事件风扇、跨节点广播（Pub/Sub） |
| 状态快照与预热缓存 | 高吞吐写入日志（Kafka 分区日志更合适） |
| 分布式锁（短临界区） | 强一致多写账本（共识协议的事） |

判断动作也只有一个：**升版本前先量主线程 CPU 占比**。`INFO stats` 看一眼 `instantaneous_ops_per_sec` 和主线程 CPU，如果主线程已经接近单核上限，调 `io-threads`、升 8.x 都不会让执行变快——瓶颈在"命令执行"这一条线上，不在读写。

## 8. 复盘：如果再来一次

1. **先画"哪个线程在扛"的执行边界图，再谈选型**。当初选"Redis 做缓存 + Kafka 做队列"时，如果能先把四种线程（主线程 / BIO / IO / 消费者）的执行边界画出来，会更快意识到：Redis 的吞吐承诺是"单线程执行"的承诺，Kafka 的吞吐承诺是"分区日志顺序写"的承诺，两者不是同一种快。
2. **"Redis 快 vs Kafka 可靠"是个伪对比**。真正的对比是"单线程内存点查 vs 分区日志顺序写"——前者适合状态，后者适合事件流。选错层位时，调参救不了，换位置才有效。

---

这次走读把"Redis 快"拆成了"哪种快"：点查快、状态快照快；但事件分发、高并发写回并不快——因为命令执行只有一条线程。对我来说，这足够回答最初的问题：**热路径上 Redis 该不该出现、出现在哪一层。** 升级救不了执行瓶颈，层位能。

**相关阅读（站内规划）**
- 低延迟系统 · Redis 缓存设计：热 key、区间扫描、修正失效与集群批量
- 低延迟系统 · 撮合引擎核心：订单簿结构、撮合语义与并发读取
- 低延迟系统 · Aeron 选型实录（待写）

## 附录 A：源码核对清单

正文结论 → 文件 / 函数 / tag 的对照表，方便自行打开核对：

| 结论 | 文件 / 函数 | tag |
|---|---|---|
| BIO 线程按任务类型各建一条 | `bio.c` `bioInit` / `bioProcessBackgroundJobs` | `6.0.0` |
| BIO 只做关闭文件、fsync、lazyfree | `bio.c` `bioProcessBackgroundJobs` switch | `6.0.0` |
| IO 线程只做读写 | `networking.c` `IOThreadMain` | `6.0.0` |
| 执行回主线程 | `networking.c` `handleClientsWithPendingReadsUsingThreads` | `6.0.0` |
| 默认不搬读 | `networking.c` `postponeClientRead`（受 `io_threads_do_reads` 控制） | `6.0.0` |
| 低负载回退单线程 | `networking.c` `stopThreadedIOIfNeeded` | `6.0.0` |
| 8.x 实现重写、IO 线程自带事件循环 | `iothread.c` `IOThreadMain` | `8.8.0` |
| 8.x 执行仍在主线程（注释原文） | `iothread.c` | `8.8.0` |
| 8.x 许可变更 | `iothread.c` 文件头 | `8.8.0` |

## 附录 B：全量压测数据

环境：WSL Ubuntu 24.04（28 vCPU / 15 GiB），localhost 回环；Redis tag `6.0.0` 本机编译（`MALLOC=libc`）。吞吐用官方 `redis-benchmark`（C 客户端），延迟分布用自写 RESP 客户端（单客户端模式避免客户端瓶颈）。每档数据为一次实验的原始输出。

| 实验 | 配置 | rps | p50 | p99 | p999 | 备注 |
|---|---|---|---|---|---|---|
| GET 3B，c50 n500k | io1 | 170,823 | — | 99.93% ≤ 0.8 ms | 100% ≤ 8 ms | 官方 benchmark |
| GET 3B，c50 n500k | io4 | 127,975 | — | 99.93% ≤ 0.9 ms | 99.99% ≤ 662 ms | 长尾明显 |
| GET 3B，c50 n500k | io4+reads | 166,555 | — | 99.95% ≤ 0.8 ms | 99.97% ≤ 9 ms | 与 io1 持平 |
| GET 3B，单客户端 n10k | io1 | 25,672 | 0.035 | 0.094 | 0.141 | 干净延迟 |
| GET 3B，单客户端 n10k | io4 | 23,727 | 0.038 | 0.096 | 0.159 | 干净延迟 |
| GET 3B，单客户端 n10k | io4+reads | 25,815 | 0.035 | 0.086 | 0.136 | 干净延迟 |
| GET 4KB，4 进程×10 线程 | io1 | ~69 k | 0.523 | 1.81 | 7.7 | 主线程 CPU 68.5% |
| GET 4KB，4 进程×10 线程 | io4 | ~51 k | 0.650 | 2.62 | 8.3 | 主线程 46% / IO 各 ~2% |
| GET 4KB，4 进程×10 线程 | io4+reads | ~51 k | 0.654 | 2.60 | 3.9 | 主线程 54.5% / IO 各 ~2.5% |
| GET pipeline16，c50 n500k | io1 | 2,674,481 | — | — | — | 执行饱和 |
| GET pipeline16，c50 n500k | io4 | 2,382,171 | — | — | — | 无提升 |
| GET pipeline16，c50 n500k | io4+reads | 2,551,592 | — | — | — | 无提升 |
| PUBLISH，1 订阅者+4 进程×10 线程 | io1 | ~45 k | 0.68 | 3.06 | 5.3 | 主线程 CPU 64.6% |
| PUBLISH，1 订阅者+4 进程×10 线程 | io4 | ~38 k | 0.86 | 3.60 | 5.1 | 主线程 53.9% / IO 各 ~7% |

客户端与脚本：`redis-benchmark`（官方 6.0.0）；自写 RESP 压测客户端（单客户端干净延迟 / 多进程并发），脚本与数据留存于本机实验目录。

**没有测**：跨机延迟、生产网卡排队、集群多活——以上数字不构成任何机房级延迟结论。
