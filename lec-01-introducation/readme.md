# Map-Reduce

又看了一遍 map-reduce 的论文，其实感觉都不用看了。04 年左右的新想法，现在基本上已经是日经的思路了。之前用过几个支持分布式 task 的框架，思路、设计基本上如出一辙。

仔细的看了一下论文里面讨论的比较细节的东西吧，slide 里面也有讲到，performance, fault tolerance, consistency (性能、错误处理、一致性)。印象比较深刻的就是: 

1. 遇到错误的机器 master 重新分发处理、
2. 处理多次错误的结果发给 master 重运行的时候略过
3. backups process 处理末尾的慢速机器

其余的 M、R 的分区，在数据所在机器执行 map 什么的感觉现在都已经是共识的应用了。不过论文里面没提到 master crash 掉之后的一些具体处理方式（只有 checkpoint），04 年的作者建议我们检查问题然后重新跑，看来 master 各种选举什么的算法是这之后出现的？

另外 MR 出现的大多都以 input\outputs files 形式组织的，通过 GFS 分割的文件块 16m-64m。每个 input file 至少有 3 份备份。

## 思考问答时间

```
  Q: What's a good data structure for implementing this?
```

根据现有的理解应该是一个连通图。

```
  Q: Why not stream the records to the reducer (via TCP) as they are being produced by the mappers?
```

这个不是很现实吧，这样就每个 mapper 都要附带找对应分区的 reducer 的功能，职责区分就不是很明确了。并且这样  mapper 挂了 backups、rerun 都不能执行了，做法很鸡肋。

```
How do they get good load balance?
  Critical to scaling -- bad for N-1 servers to wait for 1 to finish.
  But some tasks likely take longer than others.
  [diagram: packing variable-length tasks into workers]
  Solution: many more tasks than workers.
    Master hands out new tasks to workers who finish previous tasks.
    So no task is so big it dominates completion time (hopefully).
    So faster servers do more work than slower ones, finish abt the same time.
```

分更多的 task 来解决一些任务过大的 balance 问题。

```
Q: Why not re-start the whole job from the beginning?
```

这个问题就没明白，是说整个 MR 的流程么？整个 MR 的流程重新执行太费时间了吧，而且还可能遇到相似的情况，所以就把失败的 mapper 重新抛一次，reducer 跑完的就放着就行了。另外就是要求 map 是纯函数，这才是分布式用函数式方式能轻松重试的一个原因。基本上详细的都概括了：

```
Details of worker crash recovery:
  * Map worker crashes:
    master sees worker no longer responds to pings
    crashed worker's intermediate Map output is lost
      but is likely needed by every Reduce task!
    master re-runs, spreads tasks over other GFS replicas of input.
    some Reduce workers may already have read failed worker's intermediate data.
      here we depend on functional and deterministic Map()!
    master need not re-run Map if Reduces have fetched all intermediate data
      though then a Reduce crash would then force re-execution of failed Map
  * Reduce worker crashes.
    finshed tasks are OK -- stored in GFS, with replicas.
    master re-starts worker's unfinished tasks on other workers.
  * Reduce worker crashes in the middle of writing its output.
    GFS has atomic rename that prevents output from being visible until complete.
    so it's safe for the master to re-run the Reduce tasks somewhere else.
```



```
  MapReduce single-handedly made big cluster computation popular.
  - Not the most efficient or flexible.
  + Scales well.
  + Easy to program -- failures and data movement are hidden.
```

