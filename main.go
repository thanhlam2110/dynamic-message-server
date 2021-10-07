package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"grpc-server/proto"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	addr = ":8000"
)

type DataService struct {
}

func main() {
	startGRpcServer()
}

func startGRpcServer() {
	// listen [8000] port
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	proto.RegisterDataServiceServer(grpcServer, &DataService{})
	reflection.Register(grpcServer)

	log.Println("grpc server is running...")
	log.Println(fmt.Sprintf("listen at %s", addr))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (ds *DataService) PutData(ctx context.Context, req *proto.PutDataRequest) (*proto.PutDataRespose, error) {

	log.Println("put data request")
	reqData, err := parseAnyData(req.Type, req.Data)
	if err != nil {
		return nil, errors.New("illegal param")
	}

	log.Println(reqData)
	return &proto.PutDataRespose{
		Err:  0,
		Desc: "success",
	}, nil
}

func (ds *DataService) PutDataStream(stream proto.DataService_PutDataStreamServer) error {

	log.Println("put data stream request")

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			respData := &proto.PutDataStreamRespose{
				Err:  0,
				Desc: "Success",
			}
			return stream.SendAndClose(respData)
		}
		if err != nil {
			log.Printf("put data stream recv error. %v", err)
			return err
		}

		recvData, err := parseAnyData(msg.Type, msg.Data)
		if err != nil {
			log.Printf("put data stream parse error. %v", err)
			continue
		}
		log.Printf("recv from client: %v", recvData)
	}

}

func (ds *DataService) GetData(req *proto.GetDataRequest, stream proto.DataService_GetDataServer) error {

	log.Println("get data request")
	// TODO parse requst data like put data

	testDatas := []string{"test data 1", "test data 2", "test data 3", "test data 4"}
	for _, v := range testDatas {
		sendData := buildGetDataResp("string", []byte(v))
		if err := stream.Send(sendData); err != nil {
			return err
		}
	}
	return nil
}

// recv than send
// send and recv are extracted and put into the coroutine to make them independent
func (ds *DataService) GetDataStream(stream proto.DataService_GetDataStreamServer) error {

	log.Printf("get data Stream request")

	for {

		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		recvData, err := parseAnyData(msg.Type, msg.Data)
		if err != nil {
			log.Printf("get data stream parse error. %v", err)
			continue
		}
		log.Printf("get data stream recv: %v", recvData)

		// send after recv
		sendMsg := fmt.Sprintf("I recv msg is :%v", recvData)
		sendData := buildGetDataStreamResp("string", []byte(sendMsg))
		if err := stream.Send(sendData); err != nil {
			log.Printf("get data stream send data error. %v", err)
			continue
		}

	}
}

func parseAnyData(valueType string, anyData *anypb.Any) (interface{}, error) {

	if anyData == nil {
		return nil, errors.New("illegal param")
	}

	var err error
	var reqData interface{}
	switch valueType {
	case "int":
		reqData, err = strconv.ParseInt(string(anyData.Value), 10, 64)
	case "float":
		reqData, err = strconv.ParseFloat(string(anyData.Value), 10)
	case "json":
		err = json.Unmarshal(anyData.Value, &reqData)
	case "bool":
		reqData, err = strconv.ParseBool(string(anyData.Value))
	case "string":
		reqData = string(anyData.Value)
	default:
		reqData = string(anyData.Value)
	}

	return reqData, err
}

func buildGetDataResp(valueType string, value []byte) *proto.GetDataRespose {
	return &proto.GetDataRespose{
		Err:  0,
		Desc: "Success",
		Type: valueType,
		Data: &any.Any{Value: value},
	}
}

func buildGetDataStreamResp(valueType string, value []byte) *proto.GetDataStreamRespose {
	return &proto.GetDataStreamRespose{
		Err:  0,
		Desc: "Success",
		Type: valueType,
		Data: &any.Any{Value: value},
	}
}

func recvHandler() {

}

func sendHandler() {

}
