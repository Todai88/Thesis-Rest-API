package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/Todai88/Thesis-GRPC/proto"
	"google.golang.org/grpc"
)

const (
	address = "localhost:5000"
)

func connectUser(client pb.GRPCClient) int32 {
	stdin := bufio.NewReader(os.Stdin)
	var id int32
	fmt.Print("Enter ID (numeric): ")
	fmt.Scanf("%d", &id)
	name := "Tester"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	stdin.ReadString('\n')

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	stream, err := client.ConnectUser(ctx, &pb.User{Name: name, Ip: "192.168.0.1", Id: id})
	if err != nil {
		log.Fatalf("Could not greet: %v", err)
	}
	resp, err := stream.Recv()
	fmt.Println(resp)
	return id
}
func subscribeStream(id int32, client pb.GRPCClient) {
	stdin := bufio.NewReader(os.Stdin)
	stream, err := client.MessageUser(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	ctx := stream.Context()
	fmt.Println("In goroutine")
	for i := 1; i <= 10; i++ {
		var targetId int32
		fmt.Println("Enter target id (numeric): ")
		fmt.Scanf("%d", &targetId)
		stdin.ReadString('\n')

		client.MessageUser(ctx)
		if err := stream.Send(&pb.Message{Sender: &pb.User{Id: id, Name: "Attacker"}, Receiver: &pb.User{Id: targetId, Name: "Sender"}, Message: "Tag, you're it! :)"}); err != nil {
			log.Fatalf("can not send %v", err)
		}
	}
}

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGRPCClient(conn)
	id := connectUser(c)

	subscribeStream(id, c)

	// rand.Seed(time.Now().Unix())

	// conn, err := grpc.Dial(address, grpc.WithInsecure())
	// if err != nil {
	// 	log.Fatalf("can not connect with server %v", err)
	// }

	// // create stream
	// client := pb.NewGRPCClient(conn)
	// stream, err := client.BidiInt(context.Background())
	// if err != nil {
	// 	log.Fatalf("openn stream error %v", err)
	// }

	// var max int32
	// ctx := stream.Context()
	// done := make(chan bool)

	// go func() {
	// 	for i := 1; i <= 10; i++ {
	// 		// generate random nummber and send it to stream
	// 		rnd := int32(rand.Intn(i))
	// 		req := pb.Request{Num: rnd}
	// 		if err := stream.Send(&req); err != nil {
	// 			log.Fatalf("can not send %v", err)
	// 		}
	// 		log.Printf("%d sent", req.Num)
	// 		time.Sleep(time.Millisecond * 200)
	// 	}
	// 	if err := stream.CloseSend(); err != nil {
	// 		log.Println(err)
	// 	}
	// }()

	// // second goroutine receives data from stream
	// // and saves result in max variable
	// //
	// // if stream is finished it closes done channel
	// go func() {
	// 	for {
	// 		resp, err := stream.Recv()
	// 		if err == io.EOF {
	// 			close(done)
	// 			return
	// 		}
	// 		if err != nil {
	// 			log.Fatalf("can not receive %v", err)
	// 		}
	// 		max = resp.Result
	// 		log.Printf("new max %d received", max)
	// 	}
	// }()

	// // third goroutine closes done channel
	// // if context is done
	// go func() {
	// 	<-ctx.Done()
	// 	if err := ctx.Err(); err != nil {
	// 		log.Println(err)
	// 	}
	// 	close(done)
	// }()

	// <-done
	// log.Printf("finished with max=%d", max)
}
