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

	pb "github.com/todai88/thesis/Thesis-GRPC/proto"
	"google.golang.org/grpc"
)

const (
	address = "localhost:5000"
)

func createUser() pb.User {
	stdin := bufio.NewReader(os.Stdin)
	var id int32
	fmt.Print("Enter ID (numeric): ")
	fmt.Scanf("%d", &id)
	name := "Tester"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	stdin.ReadString('\n')
	return pb.User{Id: id, Name: name}
}
func estblishConnectionAndSendMessages(user pb.User) error {
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
		log.Fatalf("openn stream error %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// listener := make(chan pb.Message)
		// sendErrorChannel := make(chan error)
		go func() {
			for {
				msg, err := client.Recv()
				if err == io.EOF {
					return
				}
				log.Println(msg)
			}
		}()

		// recErrorChannel := make(chan error)
		// go func() {
		// 	for {
		// 		msg, err := client.Recv()
		// 		if err == io.EOF {
		// 			close(recErrorChannel)
		// 			return
		// 		}
		// 		if err != nil {
		// 			recErrorChannel <- err
		// 			return
		// 		}
		// 		log.Printf("Received a message: %s", msg)
		// 	}
		// }()
		for {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter Target Id: ")
			text, _ := reader.ReadString('\n')
			number, _ := strconv.ParseInt(text, 10, 4)
			num := int32(number)

			fmt.Println(num)

			req := pb.Message{Sender: &pb.User{Id: user.Id, Name: user.Name},
				Receiver: &pb.User{Id: num, Name: ""},
				Message:  "Attack"}
			fmt.Println(req)
			err = client.Send(&req)
		}
		select {
		case <-client.Context().Done():
			return client.Context().Err()
		}
	}
	return nil
}

func main() {
	user := createUser()
	estblishConnectionAndSendMessages(user)
}
