package nexus

import (
	"bytes"
	"fmt"
	"github.com/StephenFaust/nexus-go/nexus/common"
	"github.com/StephenFaust/nexus-go/nexus/serializer"
	"github.com/StephenFaust/noa/codec"
	"github.com/StephenFaust/noa/io"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type RpcClient struct {
	pool       *chanelPool
	serializer serializer.Serializer
	respMap    sync.Map
	seq        atomic.Int64
}

func NewClient(address string, poolMaxNum int) *RpcClient {
	client := new(RpcClient)
	gobSerializer := serializer.ProtoSerializer{}
	noaClient := io.NewClient(&nexusClientChanelHandler{client}, codec.DefaultCodec)
	pool := chanelPool{
		chanelChan: make(chan *io.Chanel),
		factory: &connFactory{
			client:  noaClient,
			address: address,
		},
		maxNum: poolMaxNum,
	}
	client.pool = &pool
	client.serializer = gobSerializer
	return client
}

type chanelPool struct {
	chanelChan chan *io.Chanel // 存放连接的通道
	factory    *connFactory    // 创建连接的工厂函数
	maxNum     int             // 连接池的最大连接数
	num        int             // 连接池中当前的连接数
	locker     sync.Mutex      // 同步锁
}

func (pool *chanelPool) getChanel() (chanel *io.Chanel, err error) {
	// 尝试从通道中获取一个连接
	select {
	case chanel = <-pool.chanelChan:
		return chanel, nil
	default:
		// 如果通道中没有可用的连接，则创建一个新连接
		pool.locker.Lock()
		if pool.num >= pool.maxNum {
			pool.locker.Unlock()
			chanel = <-pool.chanelChan
		} else {
			chanel, err = pool.factory.connect()
			if err != nil {
				pool.locker.Unlock()
				return nil, err
			}
			pool.num++
			pool.locker.Unlock()
		}
	}
	if chanel.IsActive() {
		return chanel, err
	} else {
		return pool.getChanel()
	}
}

func (pool *chanelPool) release(chanel *io.Chanel) {
	go func() {
		pool.chanelChan <- chanel
	}()
}

type connFactory struct {
	client  *io.Client
	address string
}

func (factory *connFactory) connect() (*io.Chanel, error) {
	err, chanel := factory.client.Connect(factory.address)
	if err != nil {
		return nil, err
	}
	return chanel, nil
}

type sign = chan *common.RpcResponse

func (client *RpcClient) SendReq(request *common.RpcRequest) (resp *common.RpcResponse, err error) {
	data, err := client.serializer.Serialize(request)
	if err != nil {
		return resp, err
	}
	ch := make(sign)
	client.respMap.LoadOrStore(request.Seq, ch)
	chanel, err := client.pool.getChanel()
	if err != nil {
		return resp, err
	}
	err = chanel.WriteAndFlush(bytes.NewBuffer(data))
	client.pool.release(chanel)
	if err != nil {
		return new(common.RpcResponse), err
	}
	if err != nil {
		return resp, err
	}
	return client.wait(time.Second*1, request.Seq, ch)
}

func (client *RpcClient) wait(d time.Duration, seq int64, ch sign) (resp *common.RpcResponse, err error) {
	defer client.respMap.Delete(seq)
	select {
	case resp = <-ch:
		return resp, nil
	case <-time.After(d):
		{
			return resp, fmt.Errorf("request timeout")
		}

	}
}

func (handler *nexusClientChanelHandler) handResponse(data *bytes.Buffer) {
	resp := new(common.RpcResponse)
	err := handler.client.serializer.Deserialize(data.Bytes(), resp)
	if err != nil {
		log.Println("Serialize failed,msg: ", err)
		return
	}
	if value, ok := handler.client.respMap.Load(resp.Seq); ok {
		value.(chan *common.RpcResponse) <- resp
	} else {
		log.Println("Receive message failed")
	}
}

func (client *RpcClient) Invoke(args any, serviceName string, methodName string, reply any) error {
	if client == nil {
		return fmt.Errorf(fmt.Sprintf("%s client not initialized", serviceName))
	}
	anyArgs, err := anypb.New(args.(proto.Message))
	if err != nil {
		return err
	}
	request := common.RpcRequest{
		ServiceName: serviceName,
		MethodName:  methodName,
		Parameters:  anyArgs,
		Headers:     nil,
		Seq:         client.seq.Add(1),
	}
	response, err := client.SendReq(&request)
	if err != nil {
		return err
	}
	if response.Status == "Success" {
		err := response.Result.UnmarshalTo(reply.(proto.Message))
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf(response.ErrorMsg)
	}
	return nil
}

type nexusClientChanelHandler struct {
	client *RpcClient
}

func (handler *nexusClientChanelHandler) OnActive(chanel *io.Chanel) {
	log.Println("Channel active:", chanel)
}

func (handler *nexusClientChanelHandler) OnMessage(chanel *io.Chanel, data *bytes.Buffer) {
	go handler.handResponse(data)
}

func (handler *nexusClientChanelHandler) OnError(chanel *io.Chanel, err error) {
	log.Println(err.Error())
}

func (handler *nexusClientChanelHandler) OnClose() {
	log.Println("Connection is closed")
}
