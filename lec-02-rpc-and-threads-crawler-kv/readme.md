# lec2 RPC and Threads, Crawler, KV

讲了下为啥不用 CPP 换成 go 了，其实感觉都差不多，可能开发体验稍有提升吧。

读了下这个：

```
https://golang.org/doc/effective_go.html
```

很简单基本上是一些 golang 的最佳实践。

## Threads.

```
Threading challenges:
  sharing data 
    one thread reads data that another thread is changing?
    e.g. two threads do count = count + 1
    this is a "race" -- and is usually a bug
    -> use Mutexes (or other synchronization)
    -> or avoid sharing
  coordination between threads
    how to wait for all Map threads to finish?
    -> use Go channels or WaitGroup
  granularity of concurrency
    coarse-grained -> simple, but little concurrency/parallelism
    fine-grained -> more concurrency, more races and deadlocks
```

线程挑战的日经问题，协同数据、共享内存，粗细粒度的并行方案。这节之中提到的 Crawler 应该是一个搜索引擎的部分组件：

```
What is a crawler?
  goal is to fetch all web pages, e.g. to feed to an indexer
  web pages form a graph
  multiple links to each page
  graph has cycles // web pages 有环

Crawler challenges
  Arrange for I/O concurrency
    Fetch many URLs at the same time
    To increase URLs fetched per second
    Since network latency is much more of a limit than network capacity
  Fetch each URL only *once* // 避免浪费带宽 
    avoid wasting network bandwidth
    be nice to remote servers
    => Need to remember which URLs visited 
  Know when finished
```

典型的数据竞争 (Typical Data Races [¶](https://golang.org/doc/articles/race_detector.html#Typical_Data_Races) )：

1. Race on loop counter [¶](https://golang.org/doc/articles/race_detector.html#Race_on_loop_counter) 循环的 goroutine 方法没带参数
2. Accidentally shared variable [¶](https://golang.org/doc/articles/race_detector.html#Accidentally_shared_variable) 前后使用了相同的变量名
3. Unprotected global variable [¶](https://golang.org/doc/articles/race_detector.html#Unprotected_global_variable) 多处 goroutine 竞争同一个全局变量
4. Primitive unprotected variable [¶](https://golang.org/doc/articles/race_detector.html#Primitive_unprotected_variable) 和 Java 类似提供原子级别的 primitive 类型操作

## RPC

```
Remote Procedure Call (RPC)
  a key piece of distributed system machinery; all the labs use RPC
  goal: easy-to-program client/server communication

RPC message diagram:
  Client             Server
    request--->
       <---response
```

RPC 对失败的处理方式：

- 尽力而为
- 最多一次

```
  idea: server RPC code detects duplicate requests
    returns previous reply instead of re-running handler
```

```
Q : how to ensure XID is unique? 
```

A：设计全局统一的 key 生成。(这个记得有一致性 hash 方法)。

```
Q : server must eventually discard info about old RPCs
    when is discard safe?
```

A : 带有前一个请求的序号、支持有限次数的尝试 

```
Q : how to handle dup req while original is still executing?
```

A : 直接不处理 + pending flag 

```
What if an at-most-once server crashes and re-starts?
  if at-most-once duplicate info in memory, server will forget
    and accept duplicate requests after re-start
  maybe it should write the duplicate info to disk
  maybe replica server should also replicate duplicate info
```

A : 固定的 interval 写入硬盘和数据备份 = = 