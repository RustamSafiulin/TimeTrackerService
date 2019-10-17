package main

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/RustamSafiulin/TimeTrackerService/mail_service/api"
	"github.com/RustamSafiulin/TimeTrackerService/pkg/mail_sender"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var wg sync.WaitGroup
var jobQueue SendMailJobQueue

type server struct{}

type SendMailJob struct {
	Body string
}

type SendMailJobQueue struct {
	mailJobChan chan SendMailJob
}

func (jq *SendMailJobQueue) AddTask(mailJob SendMailJob) {
	jq.mailJobChan <- mailJob
}

func (jq *SendMailJobQueue) RunLoop() {

	go func() {
		defer wg.Done()

		for job := range jq.mailJobChan {
			//send mail
			log.Println(job.Body)
			mail_sender.Send()
		}
	}()
}

func (s *server) SendMail(ctx context.Context, r *api.SendMailRequest) (*api.SendMailResponse, error) {

	jobQueue.AddTask(SendMailJob{Body: r.Body})

	result := &api.SendMailResponse{}
	result.SendStatus = api.SendMailStatus_MailQueuedSuccess

	return result, nil
}

func main() {

	wg.Add(1)
	jobQueue.mailJobChan = make(chan SendMailJob, 200)
	jobQueue.RunLoop()

	accepter, err := net.Listen("tcp", ":3001")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	api.RegisterMailServiceServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(accepter); err != nil {
		log.Fatalf("Failed to server: %v", err)
	}

	close(jobQueue.mailJobChan)
	wg.Wait()
}
