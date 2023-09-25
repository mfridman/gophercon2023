package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"buf.build/gen/go/mfridman/gophercon2023/connectrpc/go/petstore/v1/petstorev1connect"
	petstorev1 "buf.build/gen/go/mfridman/gophercon2023/protocolbuffers/go/petstore/v1"
	"connectrpc.com/connect"
)

func main() {
	c := petstorev1connect.NewPetStoreServiceClient(http.DefaultClient, "http://localhost:8080")
	resp, err := c.ListPets(
		context.Background(),
		connect.NewRequest(&petstorev1.ListPetsRequest{}),
	)
	if err != nil {
		log.Fatal(err)
	}
	for _, pet := range resp.Msg.GetPets() {
		fmt.Println(">>>", pet.Name)
	}
}
