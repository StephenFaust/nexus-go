package test

import (
	"log"
	"nexus-go/nexus"
	"nexus-go/nexus/test/data/message"
	"testing"
)

func Test_Server(t *testing.T) {

	err := nexus.StartServer(10086,
		message.GetTestQueryServiceInfo(),
		message.GetTestCommandServiceInfo())
	if err != nil {
		log.Fatal(err)
	}
	select {}
}
