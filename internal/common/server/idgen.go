package server

import (
	"github.com/bwmarrin/snowflake"
	"github.com/sirupsen/logrus"
	"os"
)

var node *snowflake.Node

func init() {
	hostname, err := os.Hostname()
	nodeID := Hostname2WorkerID(hostname)
	logrus.Infof("hostname: %v, node id: %v\n", hostname, nodeID)
	node, err = snowflake.NewNode(nodeID)
	if err != nil {
		logrus.Fatal(err.Error())
	}
}

func Hostname2WorkerID(hostname string) int64 {
	sum := 0
	for _, ch := range hostname {
		sum += int(ch)
	}
	return int64(sum % 1024)
}

// GenerateID 生成分布式唯一ID， 目前使用的是Snowflake算法
func GenerateID() int64 {
	return node.Generate().Int64()
}
