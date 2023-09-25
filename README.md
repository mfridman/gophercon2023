# GopherCon 2023 - protobuf workshop

## Prerequisites

- [Go][go-install]
- [buf][buf-install]

[buf-install]: https://buf.build/docs/installation
[go-install]: https://go.dev/dl/

## Workshop

In this workshop we cover some awesome technologies from [Buf](https://buf.build):

- The `buf` cli
  - Tool for working with Protobuf, drop-in replacement for `protoc`
- The BSR (Buf Schema Registry)
  - Source of truth for tracking and evolving Protobuf APIs
  - Purpose-built registry
- Connect (an RPC framework)
  - Simple, interoperable, reliable. https://connectrpc.com/
  - Compatible with gRPC

### Working with .proto files (fully-offline)

```
proto/
├── buf.lock
├── buf.yaml
└── petstore
    └── v1
        └── pet.proto
```

- `buf lint proto`
  - enforces good API design choices and structure
- `buf format -w`
  - opinionated formatter for Protobuf files
  - cannot be changed

> Gofmt's style is no one's favorite, yet gofmt is everyone's favorite.

```
git init
git add .
git commit -am "commit"
```

- `buf breaking`
  - enforces compatibility and prevents breaking changes

```
buf breaking proto --against '.git#branch=main,subdir=proto'
```

All of these commands come with sane defaults. But breaking / lint can be configured.

```
buf mod init
```

Create a "named" Buf module so it can be pushed to the BSR (Buf Schema Registry):

```
name: buf.build/mfridman/gophercon2023
```

- `buf push`
  - a module is a collection of Protobuf files that are configured, built, and versioned as a
    logical unit

The BSR has complete documentation for your Protobuf files through a browsable UI with syntax
highlighting, definitions, and references.

```
buf mod open proto
```

### Dependency management

Something that's historically difficult in Protobuf ecosystem, either git submodules or
copy/paste solutions.

Declare, resolve and use hosted BSR modules as dependencies in your projects. Simply add `deps` key
to buf.yaml file and those definitions become available.

```yaml
deps:
  - buf.build/mfridman/common
```

Then run:

```
buf mod update proto
```

Which creates a buf.lock file pinning dependencies. Now in the .proto file can import and reference
the Proto dependency definitions like normal.

```proto
import "types/v1/pet.proto";

message Pet {
  ...
  types.v1.PetType pet_type = 3;
}
```

And then update / push

```
buf mod update proto
buf push proto
```

### Local code generation (protoc-gen-go)

Add a buf.gen.yaml file

```yaml
version: v1
managed:
  enabled: true
  go_package_prefix:
    default: github.com/mfridman/gophercon2023/gen
plugins:
  - plugin: go
    out: gen
    opt:
      # The output file is placed in the same relative directory as the input file
      - paths=source_relative
```

I had to install the `protoc-gen-go` Protobuf plugin locally and added it to `$PATH`:

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

- `buf generate proto --include-imports`
  - invoke plugins to generate code

### Define an API using Connect (RPC framework)

```proto
service PetStoreService {
  rpc ListPets(ListPetsRequest) returns (ListPetsResponse);
}

message ListPetsRequest {}

message ListPetsResponse {
  repeated Pet pets = 1;
}
```

#### Remote code generation

Installing Protobuf plugins locally is a bit of a pain, instead the BSR has rich support for Remote
Code Generation with hosted plugins

- https://buf.build/plugins

```yaml
plugins:
  ...
  - plugin: buf.build/connectrpc/go
    out: gen
    opt:
      - paths=source_relative
```

Now `buf generate proto --include-imports` will send the .proto files to the BSR and code generation
will happen on Buf servers in a secure sandbox. The code generation response will be sent back and
written out to disk.

#### Implement a Connect server

It's just an `http.Handler`, so interoperates with the existing Go ecosystem really nicely.

With <40 lines of code we have a fully working Connect API, just run `go run main.go`

<details>
  <summary>Go server implementation</summary>

```go
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
```

</details>

### Make requests to API

Since Connect does content-type negotiation it's REALLY easy to debug endpoints:

```sh
echo '{}' | http POST http://localhost:8080/petstore.v1.PetStoreService/ListPets | jq
```

But wait, there's more. Check out **Buf Studio**

- `buf mod open proto`

ps. don't forget to enable CORS on your server. A popular library for this:

- github.com/rs/cors

### Simple clients (Generated SDKs)

No code generation, no plugins to install. Users just `go get` the generated assets like it's any
other package in the ecosystem.

```
buf.build/gen/go/mfridman/gophercon2023/connectrpc/go
buf.build/gen/go/mfridman/gophercon2023/protocolbuffers/go
```

- https://buf.build/gen/go
  - implements the [Go module proxy](https://go.dev/ref/mod#goproxy-protocol)
- {moduleOwner}/{moduleName}
  - the module reference from the BSR
- {pluginOwner}/{pluginName}
  - the plugin reference from [buf.build/plugins](https://buf.build/plugins)

The BSR supports:

- Go module proxy
- NPM registry
- Maven registry
- Swift registry

So, we can `go get` a generated SDK:

```sh
go get buf.build/gen/go/mfridman/gophercon2023/connectrpc/go
go get buf.build/gen/go/mfridman/gophercon2023/protocolbuffers/go

go mod tidy
```

Then import the package like normal:

```go
import (
	"buf.build/gen/go/mfridman/gophercon2023/connectrpc/go/petstore/v1/petstorev1connect"
	petstorev1 "buf.build/gen/go/mfridman/gophercon2023/protocolbuffers/go/petstore/v1"
	"connectrpc.com/connect"
)
```

And start using the Go generated code:

<details>
  <summary>Go client implementation</summary>

```go
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
```

</details>

### Bonus

- The BSR can be used as an API to fetch the schema at runtime
- Although we only mentioned Connect Go, there is also rich support for:
  - Swift clients
  - Kotlin clients
  - Connect on the Web (TS/JS)
  - Servers and clients with Node
