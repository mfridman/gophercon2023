package main

import (
	"context"
	"log"
	"net/http"

	"connectrpc.com/connect"
	"github.com/go-chi/chi/v5"
	petstorev1 "github.com/mfridman/gophercon2023/gen/petstore/v1"
	"github.com/mfridman/gophercon2023/gen/petstore/v1/petstorev1connect"
	typesv1 "github.com/mfridman/gophercon2023/gen/types/v1"
	"github.com/rs/cors"
)

func main() {
	r := chi.NewRouter()
	r.Use(cors.AllowAll().Handler)
	r.Mount(petstorev1connect.NewPetStoreServiceHandler(&petStoreService{}))
	log.Fatal(http.ListenAndServe(":8080", r))
}

var _ petstorev1connect.PetStoreServiceHandler = (*petStoreService)(nil)

type petStoreService struct{}

func (p *petStoreService) ListPets(
	ctx context.Context,
	_ *connect.Request[petstorev1.ListPetsRequest],
) (*connect.Response[petstorev1.ListPetsResponse], error) {
	resp := &petstorev1.ListPetsResponse{
		Pets: []*petstorev1.Pet{
			{Name: "Rocky", PetType: typesv1.PetType_PET_TYPE_DOG},
			{Name: "Buddy", PetType: typesv1.PetType_PET_TYPE_DOG},
			{Name: "Dante", PetType: typesv1.PetType_PET_TYPE_DOG},
		},
	}
	return connect.NewResponse(resp), nil
}
