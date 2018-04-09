package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	pb "github.com/Todai88/Thesis/Thesis-GRPC/proto"
	"google.golang.org/grpc"
)

const (
	address = "localhost:5000"
)

func createUser() pb.User {
	reader := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter ID (numeric): ")
	reader.Scan()
	id, _ := strconv.ParseInt(reader.Text(), 10, 4)
	name := "Tester"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	return pb.User{Id: int32(id), Name: name}
}
func estblishConnectionAndSendMessages(user pb.User) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}
	defer conn.Close()
	// create stream

	timeout := 5 * time.Minute
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	client, err := pb.NewGRPCClient(conn).EstablishBidiConnection(ctx)
	req := pb.Message{Sender: &pb.User{Id: user.Id, Name: user.Name}, Message: "Connection", Receiver: &pb.User{Id: user.Id, Name: user.Name}}
	client.Send(&req)
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}

	go func() {
		for {
			msg, err := client.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println("read error:", err)
				return
			}
			log.Println("read message:", msg)
		}
	}()

	reader := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter Target Id: ")
	for reader.Scan() {
		attackID, err := strconv.ParseInt(reader.Text(), 10, 4)
		if err != nil {
			log.Fatal("a number is required")
		}

		req := pb.Message{
			Sender: &pb.User{
				Id:   user.Id,
				Name: user.Name,
			},
			Receiver: &pb.User{
				Id: int32(attackID),
			},
			Message: "Attack",
		}
		log.Println("Sending:", req)
		err = client.Send(&req)
		fmt.Println("Enter Target Id: ")
	}

	log.Println(reader.Err())
}

func main() {
	user := createUser()
	estblishConnectionAndSendMessages(user)
}
