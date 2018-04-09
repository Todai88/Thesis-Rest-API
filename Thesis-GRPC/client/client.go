package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	pb "github.com/todai88/thesis/Thesis-GRPC/proto"
	"golang.org/x/net/trace"
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
	// Make sure connection closes after finishing run
	defer conn.Close()

	// create stream
	timeout := 5 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// Release memory after shutdown
	defer cancel()

	client, err := pb.NewGRPCClient(conn).EstablishBidiConnection(ctx)
	req := pb.Message{Sender: &pb.User{Id: user.Id, Name: user.Name}, Message: "Connection", Receiver: &pb.User{Id: user.Id, Name: user.Name}}
	client.Send(&req)
	if err != nil {
		logError(ctx, "Something went wrong when client attempted to open stream: %v", err)
	}

	go func() {
		for {
			msg, err := client.Recv()
			if err != nil {
				if err == io.EOF {
					fmt.Println("EOF reached, exeting reader.")
					return
				}
				logError(ctx, "Something went wrong when reading: %v", err)
				return
			}
			log.Println("read message:", msg)
		}
	}()

	for {
		in := printAndRead(ctx, "Enter Target Id: ")
		if in == "q" {
			if err := client.CloseSend(); err != nil {
				logError(ctx, "CloseSend returned error in client: %v", err)
			}
			break
		}

		attackID, err := strconv.ParseInt(in, 10, 4)

		if err != nil {
			logError(ctx, "An error occurred, a number is required: %v", err)
			continue
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
		err = client.Send(&req)
		if err != nil {
			logError(ctx, "Something went wrong when client send: %v", err)
		}
	}
}

func logError(parentContext context.Context, format string, a ...interface{}) {
	tr, ok := trace.FromContext(parentContext)
	if !ok {
		return
	}
	tr.LazyPrintf(format, a...)
}

func printAndRead(parentContext context.Context, output string) string {
	reader := bufio.NewScanner(os.Stdin)
	if err := reader.Err(); err != nil {
		logError(parentContext, "Something went wrong when starting the scanner, %v", err)
	}
	fmt.Printf(output)
	reader.Scan()
	in := strings.Replace(reader.Text(), "\n", "", -1)
	return in
}
func main() {
	user := createUser()
	estblishConnectionAndSendMessages(user)
}
