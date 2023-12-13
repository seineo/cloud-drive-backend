package kafka

import (
	"common/eventbus/account"
	"crypto/tls"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"log"
	"os"
	"testing"
	"time"
)

type KafkaSuite struct {
	suite.Suite
	dialer *kafka.Dialer
}

func TestKafkaSuite(t *testing.T) {
	suite.Run(t, new(KafkaSuite))
}
func (suite *KafkaSuite) SetupSuite() {
	// 环境变量可以通过IDE设置
	userName := os.Getenv("KAFKA_USERNAME")
	password := os.Getenv("KAFKA_PASSWORD")
	assert.NotEmpty(suite.T(), userName)
	assert.NotEmpty(suite.T(), password)
	mechanism, err := scram.Mechanism(scram.SHA256, userName, password)
	if err != nil {
		log.Fatalln(err)
	}
	suite.dialer = &kafka.Dialer{
		SASLMechanism: mechanism,
		TLS:           &tls.Config{},
	}
}

func (suite *KafkaSuite) TearDownSuite() {
}

func (suite *KafkaSuite) codeHandler(eventBytes []byte, eventData map[string]interface{}) {
	eventName, exists := eventData["eventName"]
	if !exists {
		suite.T().Error("eventName is not in event")
		return
	}
	suite.T().Logf("event comes: %v", eventName)

	assert.Equal(suite.T(), "codeGenerated", eventName)

	codeEvent := account.CodeGenerated{}
	err := json.Unmarshal(eventBytes, &codeEvent)
	if err != nil {
		suite.T().Error("unable to unmarshal event to codeGenerated")
		return
	}
	assert.Equal(suite.T(), "1@test.com", codeEvent.Email)
	assert.Equal(suite.T(), "123456", codeEvent.Code)
}

func (suite *KafkaSuite) TestCloseReader() {
	producer := NewEventProducer(suite.dialer, []string{"factual-marmot-8450-us1-kafka.upstash.io:9092"})
	codeGeneratedEvent := account.NewCodeGeneratedEvent("1@test.com", "123456")
	err := producer.Publish("account", codeGeneratedEvent)
	assert.NoError(suite.T(), err)

	done := make(chan bool)

	consumer := NewEventConsumer(suite.dialer, []string{"factual-marmot-8450-us1-kafka.upstash.io:9092"})

	consumer.Subscribe("account", suite.codeHandler)
	// 开始消费
	go func() {
		err := consumer.StartConsuming("account", time.Now().Add(-5*time.Minute))
		assert.Error(suite.T(), err, io.EOF)
		done <- true
	}()
	// 定时关闭读取
	go func() {
		select {
		case <-time.After(3 * time.Second):
			err := consumer.Stop()
			if err != nil {
				suite.T().Errorf("stop consuming error: %v", err.Error())
				return
			}
		}
	}()
	<-done
}

//
//func (suite *KafkaSuite) TestTimeout() {
//	producer := NewEventProducer(suite.dialer, []string{"factual-marmot-8450-us1-kafka.upstash.io:9092"})
//	err := producer.Publish("account", account.NewCodeGeneratedEvent("1@test.com", "123456"))
//	if err != nil {
//		suite.T().Errorf(err.Error())
//	}
//
//	done := make(chan bool)
//
//	consumer := NewEventConsumer(suite.dialer, []string{"factual-marmot-8450-us1-kafka.upstash.io:9092"})
//
//	consumer.Subscribe("account", suite.codeHandler)
//	// 开始消费
//	go func() {
//		err := consumer.StartConsuming("account", time.Now().Add(-5*time.Minute))
//		assert.Error(suite.T(), err, context.DeadlineExceeded)
//		done <- true
//	}()
//	// 定时关闭读取
//	go func() {
//		select {
//		case <-time.After(10 * time.Second):
//			err := consumer.Stop()
//			if err != nil {
//				suite.T().Errorf("stop consuming error: %v", err.Error())
//				return
//			}
//		}
//	}()
//	<-done
//}
