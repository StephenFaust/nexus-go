package test

import (
	"github.com/StephenFaust/nexus-go/nexus"
	"github.com/StephenFaust/nexus-go/nexus/test/data/message"
	"log"
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
