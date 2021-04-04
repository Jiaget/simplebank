## motivation of gRPC(意义)

- 不同编程语言的通信
    - 维持前后端的通信
    - 微服务可能是用多种语言共同编写完成的
    - 一个一致的信息交换API协议
        - 通信通道（communication channel）：REST,SOAP,message queue
        - 认证机制（authentication mechanism）:Basic, OAuth, JWT
        - 负载格式（payload format）：JSON, XML, binary
        - 数据模式（Data model）
        - 错误处理(Error handling)
    
- 高效地通信</br>
    即 轻量、快速
    - 微服务之间的数据交换量之庞大
    - 移动网络受到有限带宽的限制
    
- 简单地通信</br>
    - 客户端和服务器处理核心的逻辑    
    - 剩下的交给框架来处理
    
## gRPC(Romote Procedure Call)