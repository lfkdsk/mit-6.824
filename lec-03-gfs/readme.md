# Lec 03 : GFS

## GFS Lecture

### 原则

- 组件失效被认为是常态事件，而不是意外事件。(强调错误处理、降低组件成本)
- 文件尺寸巨大。
- 绝大部分文件的修改是采用在文件尾部追加数据，而不是覆盖原有数据的方式。(符合实际的操作预期)。
- 应用程序和文件系统 API 的协同设计提高了整个系统的灵活性。

### 接口实现

1. 类似传统的文件系统 API
2. 提供快照和追加记录的功能，每个客户端可以原子性操作(成本很低)，可以实现多路结果合并。

### 实现方式

Master、Chunk、Client 的低耦合的实现方案，跟 MR 本身也有异曲同工之处。Files 会被分割成大小固定的 Chunk 实际存储于 Chunk 的文件系统之中，每个 Chunk 创建的时候就会有一个 Global Unique ID。可以存在多个复制，空实现默认三个。

Master 管理 Chunk 的各种信息，在固定的心跳周期内会对 Chunk Servers 进行轮训。Master 的只包含数据的 meta-data ，client 拿到具体的 meta-data 之后会和 Chunk Servers 进行直接联系。

![gfs](imgs/gfs.png)

**Search Step** : 

1. Chunk Size 固定能够根据文件找到 index 
2. Client 发送包含文件名和 Chunk Index 的请求到 Master
3. Master 回复 Chunk Handle 和 Chunk Location (多个副本)   (Client 缓存了返回的 Meta 信息)
4. Client 直接和最近的 Chunk Server 联系，并且在缓存信息过期前都不会再进行请求了

**选择 64 M 的 Chunk Size 好处都有啥**：

1. 减少 Client 和 Master 的交互，可以多次在一个 Chunk 上面工作。

2. 对同一个块进行多次交互能够保持 TCP 的连接时间，减少网络负载。

3. Master 存储的 MetaData 信息能够更少，减少 Master 的内存压力。

   缺陷：热点过载，批处理程序可能在同时都在请求同一个 Chunk 的文件块，可能的解决方案是增加复制量，或者允许 Client -> Client 的数据流动。

**MetaData** :

1. the file and chunk namespaces,  文件、Chunk 的命名空间
2. the mapping from files to chunks, 文件、Chunk 的 Mapping 关系
3. and the locations of each chunk’s replicas.  每个 Chunk 的副本位置

前两者在 Master 的 log 里面也会存一份缓存到 disk 上面，防止 Master 的意外崩溃。Master 只在内存上缓存数据，而是会启动轮询 Chunk Server 去拿相应的信息。

Master 上的 MetaData 会定期存储到文件系统上，64 bit 就能管理一个 64 M 的 Chunk ，因此 Master 的内存消耗和扩容问题都不大。

Opertation Log —— 操作日志上保存着 GFS 的重要信息在写入 disk 前对 Client 不可见，并且有多个远程备份，在错误回复阶段使用 Log 进行重做，加入 CheckPoint 的设定，能够减少需要恢复的量。压缩 B 树的文件结构能够直接映射到内存。

``` 
The state of a file region after a data mutation depends on the type of mutation, whether it succeeds or fails, and whether there are concurrent mutations. Table 1 summa- rizes the result.
```

![regison_success](imgs/region_success.png)

**概念** ：

- **consistent** : 如果所有客户端不论从哪一个备份中读取同一个文件，得到的结果都是相同的，那么我们就说这个文件空间是一致的。
- **defined：**如果一个文件区域在经过一系列操作之后依旧是一致的，并且客户端完全知晓对它所做的所有操作。
- 一个操作如果没有被其他并发的写操作影响，那么这个被操作的文件区域是 defined 的。
- 成功的并发操作也会导致文件区域 undefined，但是一定是一致的（consistent）（客户端有可能只看到了最终一致的结果，但是它并不知道过程）。
- 失败的并发操作会导致文件区域 undefined，所以一定也是不一致的（inconsistent）。
- GFS 并不需要是因为什么导致的 undefined（不区分是哪种 undefined），它只需要知道这个区域是 undefined 还是 defined 就可以。

**数据改变**：

- **write** ：往应用程序指定的 offset 进行写入
- **record append** ：往并发操作进行过的 offset 处进行写入，这个 offset 是由 GFS 决定的（至于如何决定的后面会有介绍），这个 offset 会作为 defined 区域的起始位置发送给 client。
- **“regular” append** ：对应于 record append 的一个概念，普通的 append 操作通常 offset 指的是文件的末尾，但是在分布式的环境中，offset 就没有这么简单了

**行为确保**：

``` 
After a sequence of successful mutations, the mutated file region is guaranteed to be defined and contain the data writ- ten by the last mutation.
```

1. GFS 通过在所有的备份（replicas）上应用顺序相同的操作来保证一个文件区域的 defined（具体细节后面会讨论）
2. GFS 会使用 chunk version（版本号）来检测 replicas 是否过期，过期的 replicas 既不会被读取也不会被写入
3. GFS 通过握手（handshakes）来检测已经宕机的 chunkserver
4. GFS 会通过校验和（checksuming）来检测文件的完整性

