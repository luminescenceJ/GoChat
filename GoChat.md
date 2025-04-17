
登录与消息流程：
用户通过HTTP进行登录，拿到jwt后访问websocket
然后在websocket下，进行token校验，然后redis记录人数
此时


分房间机制： 减少用户增删时的锁竞争

消息一致性：

消息冲消费：

消息的可靠性：

消息的持久化：

系统的并发：

消息幂等性设计：

错误方案：

锁的竞争与分片设计：

```go
// 传统方案：全局锁 + 全局用户列表
var (
    globalLock sync.RWMutex
    globalUsers map[int]*User
    globalRooms map[int]*Room
)
// 任何用户登录、退出或消息发送操作都需要竞争全局锁，高并发时锁冲突严重。
// 通过哈希将用户分配到不同的 Bucket，每个 Bucket 有自己的锁（cLock）和数据结构（chs, rooms）。
// 全局锁 → 分片锁
```



用户在线机制：Redis位图，Redis存储用户session



离线消息支持机制：

> 用户在线时接收服务器推送，离线/宕机时使用心跳机制和消息ack确保消息不丢失，然后等到客户端重新上线时主动拉去离线信息列表。



历史消息记录设计:

> 仅存储私聊记录，以接收方userId作为key，存储到redis中。对于公共频道，消息丢失后尝试3次自动丢弃





为什么选择websocket

> **错误的HTTP应用场景**
> 依赖于客户端轮询服务，而不是由用户主动发起。
> 需要频繁的服务调用来发送小消息。
> 客户端需要快速响应对资源的更改，并且，无法预测更改何时发生。
> **错误的WebSockets应用场景**
> 连接仅用于极少数事件或非常短的时间，客户端无需快速响应事件。
> 需要一次打开多个WebSockets到同一服务。
> 打开WebSocket，发送消息，然后关闭它 - 然后再重复该过程。
> 消息传递层中重新实现请求/响应模式。



```go
// Channel 结构扩展
type Channel struct {
    // ... 原有字段 ...
    pendingMessages map[string]*proto.Msg // 待确认消息（key: 消息ID）
    ackCh          chan string            // ACK 通道
    mu             sync.Mutex             // 保护 pendingMessages
}

// 发送消息时记录待确认
func (ch *Channel) SendWithACK(msg *proto.Msg) {
    ch.mu.Lock()
    defer ch.mu.Unlock()
    msg.ID = tools.GenerateUUID() // 生成唯一ID
    ch.pendingMessages[msg.ID] = msg
    ch.broadcast <- msg
}

// 在 readPump 中处理 ACK
func (s *Server) readPump(ch *Channel, c *Connect) {
    // ... 原有代码 ...
    for {
        _, message, err := ch.conn.ReadMessage()
        // ... 错误处理 ...
        
        // 解析消息类型
        var ackMsg proto.AckMessage
        if err := json.Unmarshal(message, &ackMsg); err == nil && ackMsg.Type == "ACK" {
            ch.ackCh <- ackMsg.MessageID // 处理 ACK
            continue
        }
        
        // ... 其他消息处理逻辑 ...
    }
}

// 独立协程处理 ACK 和重传
func (ch *Channel) startAckHandler() {
    ticker := time.NewTicker(5 * time.Second) // 每 5 秒检查超时
    defer ticker.Stop()
    for {
        select {
        case ackID := <-ch.ackCh:
            ch.mu.Lock()
            delete(ch.pendingMessages, ackID)
            ch.mu.Unlock()
        case <-ticker.C:
            ch.mu.Lock()
            for id, msg := range ch.pendingMessages {
                if time.Since(msg.SentTime) > 10*time.Second { // 超时重传
                    ch.broadcast <- msg
                    msg.SentTime = time.Now()
                }
            }
            ch.mu.Unlock()
        }
    }
}
```



#### 使用mongo存储消息



#### Kafka使用

安装目录 : D:\lumin\kafka_2.13-4.0.0

启动
这里只是启动单机版

格式化日志目录
首先生成一个随机的cluster.id,在命令控制台cmd上进入到目录bin\windows。

```shell
kafka-storage.bat random-uuid #然后他就会输出一个uuid。我这里是0vJqs3JPTJiq1qfd0VG4yw
```

接下来就用这个uuid作为cluster.id来格式化日志（其实就是kafka的topic数据那些）目录。

```shell
kafka-storage.bat format --standalone -t cAmL00I2SXaptS3wgPYApg -c ../../config/server.properties # 初始化完之后在日志目录E:\apps\kafka_2.13-4.0.0\data中配置好meta.properties等信息.
做完初始化就可以启动单机服务了,启动命令如下
```

```shell
kafka-server-start.bat ../../config/server.properties # 创建topic
```

检查当前 Topic：

```sh
D:\lumin\kafka_2.13-4.0.0\bin\windows>.\kafka-topics.bat --list --bootstrap-server 127.0.0.1:9092
```

Cli手动创建

```shell
kafka-topics.sh --create \
  --bootstrap-server 127.0.0.1:9092 \
  --replication-factor 1 \
  --partitions 1 \
  --topic chat_message
# k.CreateTopic() 代码中创建
# AllowAutoTopicCreation: true, 配置项自动创建
```

代码生成消息之后，手动测试消费

```shell
.\kafka-console-consumer.bat --bootstrap-server 127.0.0.1:9092 --topic chat_message --from-beginning
```





#### redis下的ack机制设计

> 问题：redis List没有acak机制，一旦消息消费后，消息的持久化就交给了服务器。但是如果服务器宕机了就会造成消息丢失。

消息队列 + 哈希表，分别表示未消费和消费中的消息。通过客户端的ack信息来改变哈希表中消息状态并且删除。

```go
// 双消息队列设计：待处理和处理中队列，待处理无条件弹出，可以进行流量限制，处理中只能由客户端ack后异步出队列。
func PopAndProcess() {
    for {
        msg, err := redisClient.RPopLPush("queue", "processing").Result()
        if err != nil {
            // 网络/连接问题
            continue
        }

        err = sendToClient(msg)
        if err == nil {
            redisClient.LRem("processing", 0, msg) // 成功确认
        } else {
            // 不确认，待下次重新投递
        }
    }
}
```



### 需求：高并发的可靠聊天室

#### 基本功能：历史消息获取，即时信息获取，私聊，群聊，人数统计，用户登录登出。

#### 可靠性分析：消息发送不丢失



#### 消息有序性：确保消息按发送顺序正确到达客户端

1.消息队列层面：

>  Kafka 本身是 **按 partition 保证有序** 的，使用相同 key 的消息进入同一 partition：
>
> ```go
> kafka.Message{
>     Key: []byte(userID),  // 同一个用户/群组/会话使用固定 key
>     Value: jsonBytes,
> }
> ```
>
> - Kafka 保证 **同一个 Partition 中消息顺序一致**（FIFO）。
> - 所以只要将同一会话/用户的消息都投递到同一 Partition，即可确保读取时有序。
> - Kafka 的 **消费者读取也是顺序的**（按 offset 拉取）
>
> ⚠ 注意：
>
> - 如果随机分配 Partition，会打乱顺序。
> - Partition 数不宜过少（影响并发），也不宜太多（增加协调复杂度）。

2. 业务层面：

> ✔ 每条消息附带时间戳 / 序列号：
>
> 在服务端收到多条消息时按时间戳/序列号排序后处理。
>
> ```json
> {
>   "sender_id": "u001",
>   "msg": "hello",
>   "seq": 1024
> }
> ```
>
> 客户端收到消息后可以校验是否有“丢帧”或“乱序”，必要时发起重传请求。
>
> 注意：分布式环境下服务器集群的时钟不一致，不能完全依靠时间戳进行顺序验证。而全局自增Id将会在分布式环境下成为性能瓶颈。参考[如何保证IM实时消息的“时序性”与“一致性”？ - 知乎](https://zhuanlan.zhihu.com/p/138563000)

3.websocket连接：

>因为每个用户的 websocket 写是通过一个 goroutine 从 channel 顺序读取的（你当前架构），只要消息进入 channel 是有序的，写出的顺序就是有序的。

#### 消息流程

消息发送逻辑：首先通过websocket维持连接，从chan中读取信息加入kafka队列，由kafka进行实时推送，拉取出的消息存储到mongo中进行持久化，当成功写入时更新状态。

我目前的设计采用了聊天室分桶减少锁竞争，然后希望将所有在线用户进行websocket连接（http访问页面登录，然后连接），我只需要对websocket的读写通道进行维护就可以与客户端交互。 然后我使用了chan *message 作为用户接收消息的通道，websocket去监听这个通道，有数据就传到客户端，同样的。我的客户端通过通道发送消息到websocket。我目前的设计是采用一个任务层+消息队列来实现，比如客户端发来消息之后，我直接将消息加入消息队列kafka，然后由任务层持续对kafka进行消费，比如拿取信息之后发现是群发，就会调用群发的api，枚举所有chan *message发送，然后websocket再进行监听，发送到客户端。 你觉得这种架构怎么样？有哪些可以改进的地方和不需要的地方。我的需求是高并发高性能

> **分桶(bucket)减少锁竞争**
>
> **WebSocket长连接做消息推送**
>
> **chan \*message 作为用户的接收通道**
>
> **Kafka 做异步解耦，消息任务层处理业务逻辑**

> **问题：** 每个用户一个 goroutine 监听 chan → WebSocket，数量大时 goroutine 激增。
>
> **问题：** 用户通道消息未消费时可能堆积。
>
> **问题：** Kafka 默认 `acks=0`，容易丢消息。
>
> **建议**：对 WebSocket 连接数、消息写入 Kafka、消费者消费延迟做 metrics 监控。