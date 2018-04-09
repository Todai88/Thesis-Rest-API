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

	pb "github.com/Todai88/Thesis/Thesis-GRPC/proto"
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
	channel.listenerMu.Lock()
	defer channel.listenerMu.Unlock()
	if c, ok := channel.listeners[id]; ok {
		close(c)
		delete(channel.listeners, id)
	}
}

func (channel *MessageChannel) SendMessage(ctx context.Context, msg pb.Message) {
	channel.listenerMu.RLock()
	receiver := msg.Receiver
	fmt.Println(msg)
	defer channel.listenerMu.RUnlock()
	listener, ok := channel.listeners[receiver.Id]
	if !ok {
		panic("no such listener")
	}
	fmt.Println("Reciever: ", receiver.Id)
	fmt.Println(listener)
	select {
	case listener <- msg:
	case <-ctx.Done():
	}
}

func (channel *MessageChannel) Broadcast(ctx context.Context, msg pb.Message) {
	channel.listenerMu.RLock()
	defer channel.listenerMu.RUnlock()
	for _, listener := range channel.listeners {
		select {
		case listener <- msg:
		case <-ctx.Done():
			return
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
	fmt.Printf("A new user just with id %d connected: %s. Now we have: %d active users\n%v\n", sender.Id, sender.Name, len(s.channels.listeners)+1, s.channels.listeners)
	s.channels.Broadcast(stream.Context(), pb.Message{Sender: sender, Message: "Connected"})

	err = s.channels.Add(sender.Id, listener)
	if err != nil {
		return err
	}
	defer func(sender *pb.User) {
		s.channels.Remove(sender.Id)
		fmt.Printf("%s has left the channel", sender.Name)
		s.channels.Broadcast(stream.Context(), pb.Message{Sender: sender, Message: "Disconnected"})
	}(sender)

	sendErrorChannel := make(chan error)
	go func() {
		for {
			select {
			case msg, ok := <-listener:
				fmt.Println(msg, ok)
				if !ok {
					return
				}
				err = stream.Send(&msg)
				if err != nil {
					sendErrorChannel <- err
					return
				}
			case <-stream.Context().Done():
				return
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
			return nil
		}
		return err
	case err := <-sendErrorChannel:
		return err
	case <-stream.Context().Done():
		return stream.Context().Err()
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
