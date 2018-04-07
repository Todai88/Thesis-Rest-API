package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	pb "github.com/Todai88/Thesis-GRPC/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type User struct {
	name, ip string
	id       int
	stream   pb.GRPC_ConnectUserServer
}

type Msg struct {
	message string
}

type server struct {
	users map[int32]User
}

func (s *server) ConnectUser(in *pb.User, stream pb.GRPC_ConnectUserServer) error {
	s.users[in.Id] = User{id: int(in.Id), name: in.Name, ip: in.Ip, stream: stream}
	alert := pb.Message{Message: "You are now connected."}
	stream.Send(&alert)
	fmt.Printf("A new user just with id %d connected: %s. Now we have: %d\n", in.Id, in.Name, len(s.users))
	for {

	}
	return nil
}

func (s *server) MessageUser(stream pb.GRPC_MessageUserServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("exit")
			return nil
		}
		fmt.Println(req)
		targetId := req.Receiver.Id
		target, ok := s.users[targetId]
		if !ok {
			return errors.New("Couldn't find a user with that ID")
		}

		attackerId := req.Sender.Id
		sender, ok := s.users[attackerId]
		if !ok {
			return errors.New("Couldn't find a user with that ID")
		}

		target.stream.Send(&pb.Message{Sender: &pb.User{Id: attackerId, Name: sender.name}, Receiver: &pb.User{Id: targetId, Name: target.name}, Message: "Tag, you're it! :)"})
		return nil
	}
	return errors.New("Couldn't find user with that id")
}

func (s *server) pushMessages() {

}

func newServer() *server {
	return &server{users: make(map[int32]User)}
}

func main() {
	// users = make(map[int32]User)
	lis, err := net.Listen("tcp", ":5000")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	myServer := newServer()

	pb.RegisterGRPCServer(s, myServer)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *server) BidiInt(srv pb.GRPC_BidiIntServer) error {
	log.Println("Started new server")
	var max int32
	ctx := srv.Context()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := srv.Recv()
		if err == io.EOF {
			log.Println("exit")
			return nil
		}

		if err != nil {
			log.Printf("Received an error: %v", err)
			continue
		}

		if req.Num <= max {
			continue
		}

		max = req.Num
		resp := pb.Response{Result: max}

		if err := srv.Send(&resp); err != nil {
			log.Printf("Send error: %v", err)
		}

		log.Printf("Send new max: %d", max)
	}
}
