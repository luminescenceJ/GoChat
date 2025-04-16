package tools

import (
	"Go-Chat/config"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"time"
)

type KafkaService struct {
	ChatWriter *kafka.Writer
	ChatReader *kafka.Reader
	KafkaConn  *kafka.Conn
}

// KafkaInit 初始化kafka
func (k *KafkaService) KafkaInit(kafkaConfig config.KafkaConfig) {

	//k.CreateTopic()
	a := kafka.TCP(kafkaConfig.HostPort)
	fmt.Println(a)
	b := &kafka.Hash{}
	fmt.Println(b)

	fmt.Printf("kafkaConfig.HostPort: %T = %v\n", kafkaConfig.HostPort, kafkaConfig.HostPort)

	k.ChatWriter = &kafka.Writer{
		Addr:                   kafka.TCP(kafkaConfig.HostPort),
		Topic:                  kafkaConfig.ChatTopic,
		Balancer:               &kafka.Hash{},
		WriteTimeout:           kafkaConfig.Timeout * time.Second,
		RequiredAcks:           kafka.RequireNone,
		AllowAutoTopicCreation: false,
	}
	k.ChatReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaConfig.HostPort},
		Topic:          kafkaConfig.ChatTopic,
		CommitInterval: kafkaConfig.Timeout * time.Second,
		GroupID:        "chat",
		StartOffset:    kafka.LastOffset,
	})

}

func (k *KafkaService) KafkaClose() {
	if err := k.ChatWriter.Close(); err != nil {
		logrus.Errorf("Kafka Close Error!")
	}
	if err := k.ChatReader.Close(); err != nil {
		logrus.Errorf("Kafka Close Error!")
	}
}

// CreateTopic 创建topic
func (k *KafkaService) CreateTopic() {
	// 如果已经有topic了，就不创建了
	kafkaConfig := config.Conf.Kafka

	chatTopic := kafkaConfig.ChatTopic

	// 连接至任意kafka节点
	var err error
	k.KafkaConn, err = kafka.Dial("tcp", kafkaConfig.HostPort)
	if err != nil {
		logrus.Errorf(err.Error())
	}

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             chatTopic,
			NumPartitions:     kafkaConfig.Partition,
			ReplicationFactor: 1,
		},
	}

	// 创建topic
	if err = k.KafkaConn.CreateTopics(topicConfigs...); err != nil {
		logrus.Errorf(err.Error())
	}

}

func GetKafkaInstance(kafkaConfig config.KafkaConfig) *KafkaService {
	Service := new(KafkaService)
	Service.KafkaInit(kafkaConfig)
	return Service
}
