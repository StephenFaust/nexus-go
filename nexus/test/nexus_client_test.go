package test

import (
	"log"
	"nexus-go/nexus"
	"nexus-go/nexus/test/data/message"
	"testing"
	"time"
)

func Test_Client(t *testing.T) {
	req := message.TestRequest{
		A: 1,
		B: 2,
	}

	userReq := message.UserReq{
		Id: 1,
	}

	listUserReq := message.ListUserReq{
		Ids: []int64{1, 2, 3},
	}
	client := nexus.NewClient("127.0.0.1:10086", 4)
	ServiceInstance := message.NewTestCommandServiceClient(client)

	serviceInstance2 := message.NewTestQueryServiceClient(client)

	for true {

		go func() {
			user := new(message.User)
			err := serviceInstance2.GetUserById(&userReq, user)
			if err != nil {
				log.Println(err.Error())
			} else {
				log.Println("the add answer is ", user)
			}

			userResp := new(message.UserResp)
			err = serviceInstance2.ListUserByIds(&listUserReq, userResp)
			if err != nil {
				log.Println(err.Error())
			} else {
				log.Println("the add answer is ", userResp.Users)
			}
		}()

		for i := 0; i < 10; i++ {
			go func() {
				resp := new(message.TestResponse)
				err := ServiceInstance.Add(&req, resp)
				if err != nil {
					log.Println(err.Error())
				} else {
					log.Println("the add answer is ", resp.C)
				}
			}()
		}
		time.Sleep(5 * time.Second)
	}

}
