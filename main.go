package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/anthdm/foreverstore/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcptransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)

	fileServerOpts := FileServerOpts{
		EncKey:            newEncryptionKey(),
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	s := NewFileServer(fileServerOpts)

	tcpTransport.OnPeer = s.OnPeer

	return s
}

func main() {
	 // 1. Get the address to listen on. 
    // In GKE, we usually listen on all interfaces (":3000")
    listenAddr := os.Getenv("LISTEN_ADDR")
    if listenAddr == "" {
        listenAddr = ":3000"
    }

    // 2. Get the list of peers to connect to.
    // Example: "vault-0.storage-service:3000,vault-1.storage-service:3000"
    bootstrapNodesStr := os.Getenv("BOOTSTRAP_NODES")
    var bootstrapNodes []string
    if bootstrapNodesStr != "" {
        bootstrapNodes = strings.Split(bootstrapNodesStr, ",")
    }

    // 3. Initialize the server using your existing makeServer function
    s := makeServer(listenAddr, bootstrapNodes...)

    fmt.Printf("Starting StorageVault node on %s\n", listenAddr)
    if len(bootstrapNodes) > 0 {
        fmt.Printf("Attempting to peer with: %v\n", bootstrapNodes)
    }

    // 4. Start the blocking server
    log.Fatal(s.Start())

 
}
