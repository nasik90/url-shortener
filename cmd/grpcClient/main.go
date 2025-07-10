package main

import (
	"context"
	"fmt"

	pb "github.com/nasik90/url-shortener/internal/app/grpcapi"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	// получаем переменную интерфейсного типа UsersClient,
	// через которую будем отправлять сообщения
	c := pb.NewShortenerClient(conn)

	md := metadata.Pairs("userId", "grpcUser1", "X-Real-IP", "192.168.1.100")
	//md := metadata.Pairs("userId", "grpcUser1")
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// функции, в которых будем отправлять сообщения
	getShortURL(ctx, c)
	getShortURLs(ctx, c)
	getUserURLs(ctx, c)
	markRecordsForDeletion(ctx, c)
	GetURLsStats(ctx, c)
}

func getShortURL(ctx context.Context, c pb.ShortenerClient) {
	var req pb.GetShortURLRequest
	req.OriginalURL = "grpc1.ru"

	resp, err := c.GetShortURL(ctx, &req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp.ShortURL)
}

func getShortURLs(ctx context.Context, c pb.ShortenerClient) {
	var originalURLWithID pb.OriginalURLWithID
	originalURLWithID.OriginalURL = "getShortURLs.ru"
	originalURLWithID.CorrelationID = "1"

	var req pb.GetShortURLsRequest
	req.OriginalURLs = append(req.OriginalURLs, &originalURLWithID)
	resp, err := c.GetShortURLs(ctx, &req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp.ShortURLs)
}

func getUserURLs(ctx context.Context, c pb.ShortenerClient) {
	resp, err := c.GetUserURLs(ctx, nil)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp.ShortOriginalURLs)
}

func markRecordsForDeletion(ctx context.Context, c pb.ShortenerClient) {
	var req pb.MarkRecordsForDeletionRequest
	req.ShortURLs = append(req.ShortURLs, "2n5VfpPt")
	c.MarkRecordsForDeletion(ctx, &req)
}

func GetURLsStats(ctx context.Context, c pb.ShortenerClient) {
	resp, err := c.GetURLsStats(ctx, nil)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)
}
