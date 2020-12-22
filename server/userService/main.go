package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	pb "github.com/PawelKowalski99/gardener_project/grpc/proto/user_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	port = flag.Int("port", 10000, "The server port")
)

type userServiceServer struct {
	pb.UnimplementedUserServiceServer
	mu sync.Mutex

	users map[int32]*pb.User
}

func newServer(options ...func(*userServiceServer)) *userServiceServer {
	s := userServiceServer{
		users: make(map[int32]*pb.User),
	}
	return &s
}

func (s *userServiceServer) CreateUser(ctx context.Context, user *pb.User) (*pb.User, error) {
	log.Printf("Create User method called")

	s.users[user.Id.Id] = user

	log.Printf("Created user: %v", user)

	return user, nil
}

func (s *userServiceServer) GetUser(ctx context.Context, id *pb.ID) (*pb.User, error) {
	log.Printf("Get User method called")
	if s.users[id.Id] == nil {
		s := "There is no user with id:" + fmt.Sprint(id.Id)
		return nil, status.New(codes.NotFound, s).Err()
	}
	log.Printf("Got user %v", s.users[id.Id])
	return s.users[id.Id], nil
}

func (s *userServiceServer) GetUsers(stream pb.UserService_GetUsersServer) error {
	log.Printf("GetUsers method called")
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		s.mu.Lock()
		var user *pb.User
		user = s.users[in.Id]
		if user == nil {
			user = &pb.User{}
		}
		s.mu.Unlock()

		if err := stream.Send(user); err != nil {
			return err

		}
	}
}

func (s *userServiceServer) DeleteUser(ctx context.Context, id *pb.ID) (*pb.ID, error) {
	log.Printf("Deleting user with id: %v", id)

	s.mu.Lock()
	delete(s.users, id.Id)
	s.mu.Unlock()

	log.Printf("Deleted user with id: %v", id)

	return id, nil
}

//   func (s *UserServiceServer) UpdateUser(ctx context.Context update *pb.UpdateUserRequest) (*pb.User, error) {
// 	return &User{
// 		Id: 3,
// 		Name: "Pawel3",
// 	}, nil
//   }

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
