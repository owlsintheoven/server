module owlsintheoven/learning-go/redis

go 1.20

replace owlsintheoven/learning-go/common => ../common

require (
	github.com/redis/go-redis/v9 v9.5.1
	golang.org/x/net v0.22.0
	owlsintheoven/learning-go/common v0.0.0-00010101000000-000000000000
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)
