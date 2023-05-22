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
	"reflect"
	"sync"
)

type RpcServer struct {
	serviceMap sync.Map
	serializer serializer.Serializer
}

func StartServer(port int, serviceInfos ...ServiceInfo) error {
	server := new(RpcServer)
	for _, info := range serviceInfos {
		server.Registry(info)
	}
	return server.start(port)
}

type ServiceInfo struct {
	Name    string
	Value   reflect.Value
	Methods map[string]*MethodInfo
}

type MethodInfo struct {
	Method   reflect.Method
	ParType  reflect.Type
	RelyType reflect.Type
}

func (s *RpcServer) Registry(sr ServiceInfo) {
	s.serviceMap.LoadOrStore(sr.Name, sr)
}

type nexusServerChanelHandler struct {
	server *RpcServer
}

func (handler nexusServerChanelHandler) OnActive(chanel *io.Chanel) {
	log.Println("Channel active:", chanel)
}

func (handler nexusServerChanelHandler) OnMessage(chanel *io.Chanel, data *bytes.Buffer) {
	go handler.handleRequest(data, chanel)
}

func (handler nexusServerChanelHandler) OnError(chanel *io.Chanel, err error) {
	log.Println(err.Error())
}

func (handler nexusServerChanelHandler) OnClose() {
	log.Println("Connection is closed")
}

func (s *RpcServer) start(port int) error {
	s.serializer = serializer.ProtoSerializer{}
	server := io.NewServer(nexusServerChanelHandler{s}, codec.DefaultCodec)
	err := server.Listen(port)
	if err != nil {
		return err
	}
	log.Println("Nexus server started,port:", port)
	return nil
}

func (handler nexusServerChanelHandler) handleRequest(data *bytes.Buffer, chanel *io.Chanel) {
	req := new(common.RpcRequest)
	err := handler.server.serializer.Deserialize(data.Bytes(), req)
	if err != nil {
		log.Println("Serialize message failed,", err)
		return
	}
	methodName := req.MethodName
	serviceName := req.ServiceName
	if serviceInfo, ok := handler.server.serviceMap.Load(serviceName); ok {
		methodInfo := serviceInfo.(ServiceInfo).Methods[methodName]
		value := serviceInfo.(ServiceInfo).Value
		var args = make([]reflect.Value, 0)
		// 汇编中类的方法调用第一个参数会把类本身指针作为第一个参数入栈, 然后再入栈其它参数,这里同cpp的汇编调用方式
		// 因为方法里可能会用到里面的字段 所以需要该结构体的内存首地址
		args = append(args, value)
		elemV := reflect.New(methodInfo.ParType.Elem())
		err := req.Parameters.UnmarshalTo(elemV.Interface().(proto.Message))
		if err != nil {
			log.Println(err)
			handler.sendResp(chanel, nil, err, req.Seq)
			return
		}
		args = append(args, elemV)
		rtv := reflect.New((methodInfo.RelyType).Elem())
		args = append(args, rtv)
		res := methodInfo.Method.Func.Call(args)
		errInterface := res[0].Interface()
		var errMsg error
		if errInterface != nil {
			errMsg = errInterface.(error)
		}
		handler.sendResp(chanel, rtv.Interface(), errMsg, req.Seq)
	} else {
		handler.sendResp(chanel, nil, fmt.Errorf("service not exist or service not registy"), req.Seq)
	}
}

func (handler nexusServerChanelHandler) sendResp(chanel *io.Chanel, returnVal any, err error, seq int64) {
	var status = "Success"
	var anyVal *anypb.Any
	var errorMsg string
	if returnVal != nil {
		anyVal, err = anypb.New(returnVal.(proto.Message))
	}
	if err != nil {
		status = "Failed"
		errorMsg = err.Error()
	}
	resp := common.RpcResponse{
		Status:   status,
		Headers:  make(map[string]string),
		Result:   anyVal,
		ErrorMsg: errorMsg,
		Seq:      seq,
	}
	data, err := handler.server.serializer.Serialize(&resp)
	if err != nil {
		log.Println("Serializer message failed,", err)
	}
	err = chanel.WriteAndFlush(bytes.NewBuffer(data))
	if err != nil {
		log.Println("Failed to write data to the channel，", err)

	}
}
