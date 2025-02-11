package main

import (
	"context"
	"fmt"
	"gochat/api"
	"google.golang.org/grpc"
	"log"
	"time"
)

func sendChat(receiverHandle api.Handle, message string) {
	receiverConnStr := fmt.Sprintf("%s:%d", receiverHandle.Host, receiverHandle.Port)

	receiverConn, err := grpc.Dial(receiverConnStr, grpc.WithInsecure())
	defer receiverConn.Close()
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
		return
	}

	chatClient := api.NewGoChatClient(receiverConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	/****
	   // THIS CODE IS FOR REFERENCE ONLY FROM THE pb PACKAGE. DO NOT UNCOMMENT
	   type api.ChatRequest struct {
			From    *api.Handle
			To      *api.Handle
			Message string
	   }
	*****/

	var req api.ChatRequest
	// TODO-WORKSHOP-STEP-8: Create req struct of type api.ChatRequest to send to client.Chat method

	_, err = chatClient.Chat(ctx, &req)
	if err != nil {
		log.Printf("ERROR: Chat(): %v", err)
		USERS.Delete(receiverHandle.Name)
	}
	return
}
