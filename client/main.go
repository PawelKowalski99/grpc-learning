package main

import (
	"flag"
	"io"
	"log"
	"time"

	pb "github.com/PawelKowalski99/gardener_project/grpc/proto/user_service"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "localhost:10000", "The server address in the format of host:port")
)

func createUser(client pb.UserServiceClient, user *pb.User) {
	log.Printf("Creating user: %s", user.Name)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client.CreateUser(ctx, user)

	log.Printf("Creating user done")

	return
}

func printUser(client pb.UserServiceClient, id *pb.ID) error {
	log.Printf("Printing user with id: %s", id)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := client.GetUser(ctx, id)
	if err != nil {
		log.Printf("%v", err)
		return err
	}

	log.Printf("User with id: %s printed: %v", id, user)

	return nil
}

func printUsers(client pb.UserServiceClient, ids []*pb.ID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.GetUsers(ctx)
	if err != nil {
		return err
	}

	waitc := make(chan struct{})

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Printf("Failed to receive user: %v", err)
			}
			log.Printf("Got user %v %v", in.Id, in.Name)
		}
	}()

	for _, id := range ids {
		if err := stream.Send(id); err != nil {
			log.Printf("Failed to send id: %v", err)
			return err
		}
	}

	stream.CloseSend()
	<-waitc

	return nil
}

func deleteUser(client pb.UserServiceClient, id *pb.ID) {
	log.Printf("Deleting user with id: %v", id)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client.DeleteUser(ctx, id)

}

func main() {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("Fail to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	createUser(client, &pb.User{
		Id:   &pb.ID{Id: 500},
		Name: "Pawel",
	})

	printUser(client, &pb.ID{Id: 1})
	printUser(client, &pb.ID{Id: 500})

	printUsers(client, []*pb.ID{
		{Id: 1},
		{Id: 2},
		{Id: 500},
	})

	deleteUser(client, &pb.ID{Id: 500})

	printUser(client, &pb.ID{Id: 500})

	return
}
