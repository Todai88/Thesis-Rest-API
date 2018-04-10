package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/todai88/thesis/Thesis-GRPC/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type User struct {
	name, ip string
	id       int
}

type MessageChannel struct {
	listenerMu sync.RWMutex
	listeners  map[int32]chan<- pb.Message
}

type Server struct {
	channels MessageChannel
}

func (channel *MessageChannel) Add(id int32, listener chan<- pb.Message) error {
	channel.listenerMu.Lock()
	defer channel.listenerMu.Unlock()
	fmt.Printf("A new user just with id %d connected. Now we have: %d active users\n", id, len(channel.listeners)+1)
	if channel.listeners == nil {
		channel.listeners = map[int32]chan<- pb.Message{}
	}
	if _, ok := channel.listeners[id]; ok {
		return status.Errorf(codes.AlreadyExists, "The id %d is already in use by another user", id)
	}
	channel.listeners[id] = listener
	return nil
}

func (channel *MessageChannel) Remove(id int32) {
	fmt.Println("Entered Remove")
	channel.listenerMu.Lock()
	defer channel.listenerMu.Unlock()
	if c, ok := channel.listeners[id]; ok {
		close(c)
		delete(channel.listeners, id)
	}
}

func (channel *MessageChannel) SendMessage(ctx context.Context, msg pb.Message) {
	channel.listenerMu.RLock()
	defer channel.listenerMu.RUnlock()
	receiver := msg.Receiver
	fmt.Println("Message: ", msg)
	listener, ok := channel.listeners[receiver.Id]
	if !ok {
		panic("no such listener")
	}
	fmt.Println("Reciever: ", receiver.Id)
	select {
	case listener <- msg:
		// case <-ctx.Done():
	}
}

func (channel *MessageChannel) Broadcast(ctx context.Context, msg pb.Message) {
	channel.listenerMu.RLock()
	defer channel.listenerMu.RUnlock()
	message := msg.GetMessage()
	fmt.Println(message)
	if message == "Disconnected" {
		fmt.Println("----- Started Disconnect Broadcast -----")
	}
	fmt.Println("Inside broadcast: ", &msg)
	fmt.Println("Available listeners: ", channel.listeners)
	index := 1
	for key, listener := range channel.listeners {
		fmt.Printf("This is iteration %d\nBroadcasting to key %v with listener %v\n", index, key, listener)
		index = index + 1
		select {
		case listener <- msg:
			fmt.Printf("Just sent to channel %v\n with key %v\n", listener, key)
			// case <-ctx.Done():
			// 	fmt.Printf("Broadcast context is apparently done on %v\n with key %v\n", listener, key)
			// 	if message == "Disconnected" {
			// 		fmt.Println("----- Finished Disconnect Broadcast -----")
			// 	}
			// 	return
		}
	}
}

func (s *Server) EstablishBidiConnection(stream pb.GRPC_EstablishBidiConnectionServer) error {
	fmt.Println("User connected")
	// Get first message
	req, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			log.Println("exit")
			return nil
		}
		return err
	}

	// Check so that sender actually is set.
	if req.Sender.Id == 0 {
		return status.Error(codes.FailedPrecondition, "Missing sender ID")
	}

	// Setup sender.
	sender := req.Sender

	listener := make(chan pb.Message)
	s.channels.Broadcast(stream.Context(), pb.Message{Sender: sender, Message: "Connected"})
	err = s.channels.Add(sender.Id, listener)

	if err != nil {
		return err
	}

	defer func(sender *pb.User) {
		fmt.Println("Entered defer: ", sender)
		s.channels.Remove(sender.Id)
		s.channels.Broadcast(stream.Context(), pb.Message{Sender: sender, Message: "Disconnected"})
		fmt.Printf("%s has left the channel", sender.Name)
	}(sender)

	sendErrorChannel := make(chan error)
	go func() {
		for {
			select {
			case msg, ok := <-listener:
				if !ok {
					return
				}
				fmt.Println("-> Sening to stream:	", &msg)
				err = stream.Send(&msg)
				if err != nil {
					fmt.Println(err)
					sendErrorChannel <- err
					return
				}
				// case <-stream.Context().Done():
				// 	return
			}
		}
	}()

	recErrorChannel := make(chan error)
	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				close(recErrorChannel)
				return
			}
			if err != nil {
				recErrorChannel <- err
				return
			}
			s.channels.SendMessage(stream.Context(), *msg)
		}
	}()

	select {
	case err, ok := <-recErrorChannel:
		if !ok {
			fmt.Println("Final Select: not OK")
			return nil
		}
		return err
	case err := <-sendErrorChannel:
		return err
		// case <-stream.Context().Done():
		// 	fmt.Println("Final Select: DONE")
		// 	return stream.Context().Err()
	}
}

func main() {
	// users = make(map[int32]User)
	lis, err := net.Listen("tcp", ":5000")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	// myServer := newServer()

	pb.RegisterGRPCServer(s, &Server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	log.Print("Serving on localhost:5000")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
