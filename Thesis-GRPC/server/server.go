package main

import (
	"fmt"
	"io"
	"log"
	"net"

	pb "github.com/todai88/thesis/Thesis-GRPC/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type User struct {
	name, ip string
	id       int
	stream   pb.GRPC_EstablishBidiConnectionServer
}

type Msg struct {
	message string
}

type server struct {
	users map[int32]User
}

// func (s *server) ConnectUser(in *pb.User, stream pb.GRPC_ConnectUserServer) error {
// 	s.users[in.Id] = User{id: int(in.Id), name: in.Name, ip: in.Ip, stream: stream}
// 	alert := pb.Message{Message: "You are now connected."}
// 	stream.Send(&alert)
// 	fmt.Printf("A new user just with id %d connected: %s. Now we have: %d\n", in.Id, in.Name, len(s.users))

// 	return nil
// }

// func (s *server) MessageUser(stream pb.GRPC_MessageUserServer) error {
// 	for {
// 		req, err := stream.Recv()
// 		if err == io.EOF {
// 			log.Println("exit")
// 			return nil
// 		}
// 		fmt.Println(req)
// 		targetId := req.Receiver.Id
// 		target, ok := s.users[targetId]
// 		if !ok {
// 			return errors.New("Couldn't find a user with that ID")
// 		}

// 		attackerId := req.Sender.Id
// 		sender, ok := s.users[attackerId]
// 		if !ok {
// 			return errors.New("Couldn't find a user with that ID")
// 		}

// 		target.stream.Send(&pb.Message{Sender: &pb.User{Id: attackerId, Name: sender.name}, Receiver: &pb.User{Id: targetId, Name: target.name}, Message: "Tag, you're it! :)"})
// 		return nil
// 	}
// 	return errors.New("Couldn't find user with that id")
// }

func createMessage(message string, sender, receiver pb.User) *pb.Message {
	return &pb.Message{Sender: &sender, Receiver: &receiver, Message: message}
}

func (s *server) EstablishBidiConnection(stream pb.GRPC_EstablishBidiConnectionServer) error {

	var max int32
	ctx := stream.Context()
	resp := new(pb.Message)
	fmt.Println("Hello")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("exit")
			return nil
		}

		if err != nil {
			log.Printf("Received an error: %v", err)
			continue
		}

		if _, ok := s.users[req.Sender.Id]; !ok {
			sender := req.Sender
			fmt.Printf("A new user just with id %d connected: %s. Now we have: %d\n", sender.Id, sender.Name, len(s.users))
			s.users[sender.Id] = User{id: int(sender.Id), name: sender.Name, stream: stream}
			resp = createMessage("Welcome.", *sender, *sender)
		}

		if req.Message == "Attack" {

		}

		fmt.Println(resp)
		for k, v := range s.users {
			fmt.Println(k, " => ", v)
		}
		fmt.Println(req.Receiver.Id)
		if err := s.users[req.Receiver.Id].stream.Send(resp); err != nil {
			log.Printf("Send error: %v", err)
		}

		log.Printf("Send new max: %d", max)
	}
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
