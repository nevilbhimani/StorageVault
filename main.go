package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

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
	// If env vars are set, run in GKE/production mode (single node)
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr != "" {
		bootstrapNodesStr := os.Getenv("BOOTSTRAP_NODES")
		var bootstrapNodes []string
		if bootstrapNodesStr != "" {
			bootstrapNodes = strings.Split(bootstrapNodesStr, ",")
		}
		s := makeServer(listenAddr, bootstrapNodes...)
		fmt.Printf("Starting StorageVault node on %s\n", listenAddr)
		log.Fatal(s.Start())
		return
	}

	// Local demo mode: spin up 3 nodes, connect them, store and retrieve a file
	s1 := makeServer(":6000")
	s2 := makeServer(":6001")
	s3 := makeServer(":6002", ":6000", ":6001")

	go func() { log.Fatal(s1.Start()) }()
	go func() { log.Fatal(s2.Start()) }()

	time.Sleep(500 * time.Millisecond)

	go func() { log.Fatal(s3.Start()) }()

	time.Sleep(500 * time.Millisecond)

	// Store a file from s3 — it writes locally and streams to s1 and s2
	key := "myprivatepicture.jpg"
	data := io.LimitReader(rand.Reader, 1<<30)
	 
	fmt.Println("=> Storing file from s3...")
	if err := s3.Store(key, data); err != nil {
		log.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)

	// Delete from s3 local disk to prove it fetches from the network
	if err := s3.store.Delete(s3.ID, key); err != nil {
		log.Fatal(err)
	}

	// Fetch the file — s3 doesn't have it locally, fetches from s1 or s2
	fmt.Println("=> Fetching file from network (s3 deleted its local copy)...")
	r, err := s3.Get(key)
	if err != nil {
		log.Fatal(err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	fmt.Printf("=> Retrieved successfully: %d bytes\n", buf.Len())

	select {}
}