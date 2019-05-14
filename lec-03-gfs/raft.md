# Lab2 ：Raft 实现

a replicated state machine protocol 复制状态机协议。

可以康康这个 Raft 的图形演示：[http://thesecretlivesofdata.com/raft/](http://thesecretlivesofdata.com/raft/)

在阅读 raft 论文之前简单写下看了这个演示的想法（相较于 GFS）：

1. master 不是指定的一个(或一群) meta-data 的机器了，引入了 leader election，GFS 里面只有 lease 有选举的过程。
2. replicas 之间的数据流动依赖 log ，这一点和 GFS 的 Chunk 里面的 control flow 很像。接到多数的 log 确认之后 leader 先写入，之后再让所有的 node 执行写入。被称作 *Log Replication* 。

Leader Election:

1. 选举过程，两个时间 election timeout (150~300ms) 通过随机时间来过滤 followers 到 candicate ，之后 vote itself 向 node 申请 votes 的投票。如果接到 vote request 的 node 没有在这轮里面投过票，就会把票投给这个申请者，并且会重设自己的 election timeout。一个 candidate 拿到大多数票的时候就会成为 leader。
2. *Append Entries* 由 leader 发送给 followers (心跳包)，followers 也会给 leader 回复相应的 Append Entries。这个过程将会持续直到一个 followers 不行。
3. 选票相同的时候会进行重选，原有的 leader 失效了之后也会进行一次重新选举。

Log Replication:

1. Log Replication 的 changes 信息修改是和 append entries 心跳包一起下发的。
2. network partition 会有多个 leader 然后就能 replication 分片了。

## 阅读 Raft 论文

Replicated state machines => Replicated Log.

优点：

1. *safety* 
2. *available* 
3. do not depend on timing to ensure the consistency
4. a minority of slow servers need not impact overall system performance.

