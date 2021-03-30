package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"imoc-product/datamodels"
	"imoc-product/services"
	"log"
	"sync"
)

// url格式  amqp://账号:密码@rabbitmq服务器地址:端口号/vhost
const MQURL = "amqp://imoocuser:imoocuser@127.0.0.1:5672/imooc"

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	// 队列名称
	QueueName string
	// 交换机
	Exchange string
	// key
	key string
	// 连接信息
	Mqurl string
	sync.Mutex
}

// 创建RabbitMQ结构体事例
func NewRabbitMQ(queueName string, exchange string, key string) *RabbitMQ {
	rabbitmq := &RabbitMQ{
		QueueName: queueName,
		Exchange:  exchange,
		key:       key,
		Mqurl:     MQURL,
	}
	var err error
	// 创建rabbitmq连接
	rabbitmq.conn, err = retryConn(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "创建链接错误")
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "获取Channels失败")
	return rabbitmq
}

func retryConn(mqUrl string) (*amqp.Connection, error) {
	return amqp.Dial(mqUrl)
}

// 断开channel和connection
func (r *RabbitMQ) Destory() {
	r.channel.Close()
	r.conn.Close()
}

// 错误处理函数
func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
		panic(fmt.Sprintf("%s,%s", message, err))
	}
}

// 简单模式Step1: 创建简单模式下rabbitmq实例
func NewRabbitMQSimple(queueName string) *RabbitMQ {
	return NewRabbitMQ(queueName, "", "")
}

// 简单模式Step2: 简单模式下生产代码
func (r *RabbitMQ) PublishSimple(message string) error {
	r.Lock()
	defer r.Unlock()
	// 1.申请队列，如果队列不存在会自动创建，如果存在则跳过创建
	// 保证队列存在，消息能发送到队列中
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		// 控制消息是否持久化
		false,
		// 是否自动删除
		false,
		// 是否具有排他性
		false,
		// 是否阻塞
		false,
		nil,
	)
	if err != nil {
		return err
	}
	// 2.发送消息到队列中
	r.channel.Publish(
		r.Exchange,
		r.QueueName,
		// 如果为true，会根据exchange类型和routkey规则，如果无法找到符合条件的队列那么会把发送的消息返回给发送者
		false,
		// 如果为true，当exchange发送消息到队列后发现队列上没有绑定消费者，则会吧消息发还给发送者。
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	return nil
}

// 简单模式Step3: 简单模式消息代码
func (r *RabbitMQ) ConsumeSimple(orderService services.IOrderService, productService services.IProductService) {
	// 1.申请队列，如果队列不存在会自动创建，如果存在则跳过创建
	// 保证队列存在，消息能发送到队列中
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		// 控制消息是否持久化
		false,
		// 是否自动删除
		false,
		// 是否具有排他性
		false,
		// 是否阻塞
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an queue")

	// 消费者流控
	r.channel.Qos(
		1,     // 当前消费者一次能接受的最大消息数量
		0,     // 服务器传递的最大容量（以八位字节为单位）
		false, // 如果设置为true 对channel可用。false为对当前队列
	)

	// 2.接收消息
	msgs, err := r.channel.Consume(
		r.QueueName,
		// 用来区分多个消费者
		"",
		// 是否自动应应答
		false,
		// 是否具有排他性
		false,
		// 如果设置为true，表示不能将同一个connection中发送的消息传递给这个connection中的消费者
		false,
		// 队列消费是否阻塞  false为阻塞
		false,
		nil,
	)
	r.failOnErr(err, "Failed to consume messages")

	forever := make(chan bool)
	// 3.启用协程处理消息
	go func() {
		for d := range msgs {
			// 实现我们要处理的逻辑函数
			message := &datamodels.Message{}
			err = json.Unmarshal([]byte(d.Body), message)
			if err != nil {
				fmt.Println(err)
			}
			// 插入订单
			fmt.Println(message)
			_, err = orderService.InsertOrderByMessage(message)
			if err != nil {
				fmt.Println(err)
			}
			// 扣除商品数量
			err = productService.SubNumberOne(message.ProductID)
			if err != nil {
				fmt.Println(err)
			}

			// 如果为true表示确认所有未确认的消息. 为false表示确认当前消息
			d.Ack(false)
		}
	}()
	log.Printf("[*] Waiting for messages, To exit pres CTRL+C")
	<-forever
}

// 订阅模式创建RabbitMQ实例
func NewRabbitMQPubSub(exchangeName string) *RabbitMQ {
	// 创建RabbitMQ实例
	return NewRabbitMQ("", exchangeName, "")
}

// 订阅模式生产
func (r *RabbitMQ) PublishPub(message string) {
	// 1.尝试创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		// 交换机类型  fanout: 广播类型
		"fanout",
		// 是否持久化
		true,
		// 是否自动删除
		false,
		// 设置为true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		// 阻塞
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	// 2.发送消息
	r.channel.Publish(
		r.Exchange,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)

}

// 订阅模式消费
func (r *RabbitMQ) ReceivePub() {
	// 1.尝试创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		// 交换机类型  fanout: 广播类型
		"fanout",
		// 是否持久化
		true,
		// 是否自动删除
		false,
		// 设置为true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		// 阻塞
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	// 2.尝试创建队列。主意：队列名称不写
	q, err := r.channel.QueueDeclare(
		"", //随机生产队列名称
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an queue")

	// 绑定队列到exchange中
	err = r.channel.QueueBind(
		q.Name,
		"", // 在pub/sub模式下，这里的key要为空
		r.Exchange,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to bind exchange")

	// 消费消息
	messages, err := r.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to consume messages")

	forever := make(chan bool)

	go func() {
		for d := range messages {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	fmt.Println("退出请按 CTRL+C")
	<-forever

}

// 路由模式创建RabbitMQ实例
func NewRabbitMQRouting(exchange string, routingKey string) *RabbitMQ {
	return NewRabbitMQ("", exchange, routingKey)
}

// 路由模式发送消息
func (r *RabbitMQ) PublishRouting(message string) {
	// 1.尝试创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		// 交换机类型
		"direct",
		// 是否持久化
		true,
		// 是否自动删除
		false,
		// 设置为true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		// 阻塞
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	// 2.发送消息
	r.channel.Publish(
		r.Exchange,
		r.key,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)

}

// 路由模式接收消息
func (r *RabbitMQ) ReceivedRouting() {
	// 1.尝试创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		// 交换机类型  fanout: 广播类型
		"direct",
		// 是否持久化
		true,
		// 是否自动删除
		false,
		// 设置为true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		// 阻塞
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	// 2.尝试创建队列。主意：队列名称不写
	q, err := r.channel.QueueDeclare(
		"", //随机生产队列名称
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an queue")

	// 绑定队列到exchange中
	err = r.channel.QueueBind(
		q.Name,
		r.key, // 在pub/sub模式下，这里的key要为空
		r.Exchange,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to bind exchange")

	// 消费消息
	messages, err := r.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to consume messages")

	forever := make(chan bool)

	go func() {
		for d := range messages {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	fmt.Println("退出请按 CTRL+C")
	<-forever
}

// 话题模式创建RabbitMQ实例
func NewRabbitMQTopic(exchange string, routingKey string) *RabbitMQ {
	return NewRabbitMQ("", exchange, routingKey)
}

// 话题模式发送消息
func (r *RabbitMQ) PublishTopic(message string) {
	// 1.尝试创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		// 交换机类型
		"topic",
		// 是否持久化
		true,
		// 是否自动删除
		false,
		// 设置为true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		// 阻塞
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	// 2.发送消息
	err = r.channel.Publish(
		r.Exchange,
		r.key,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	r.failOnErr(err, "Failed to publish message")

}

// 话题模式接收消息
// 要注意key，规则
// 其中"*"用于匹配一个单词，"#"用于匹配多个单词（可以使零个）
// 匹配 imooc.* 表示匹配 imooc.hello，但是 imooc.hello.one 需要用 imooc.# 才能匹配到
func (r *RabbitMQ) ReceivedTopic() {
	// 1.尝试创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		// 交换机类型  fanout: 广播类型
		"topic",
		// 是否持久化
		true,
		// 是否自动删除
		false,
		// 设置为true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		// 阻塞
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	// 2.尝试创建队列。主意：队列名称不写
	q, err := r.channel.QueueDeclare(
		"", //随机生产队列名称
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an queue")

	// 绑定队列到exchange中
	err = r.channel.QueueBind(
		q.Name,
		r.key, // 在pub/sub模式下，这里的key要为空
		r.Exchange,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to bind exchange")

	// 消费消息
	messages, err := r.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to consume messages")

	forever := make(chan bool)

	go func() {
		for d := range messages {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	fmt.Println("退出请按 CTRL+C")
	<-forever
}
