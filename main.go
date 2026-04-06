package main

import (
	"fmt"
	"io"
	"log"
	"os"
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
	s1 := makeServer(":3000", "")
	s2 := makeServer(":8000", "")
	s3 := makeServer(":9000", ":3000", ":8000")

	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(500 * time.Millisecond)
	go func() { log.Fatal(s2.Start()) }()

	time.Sleep(2 * time.Second)

	go s3.Start()
	time.Sleep(2 * time.Second)
	
	fileName := "test_big_file.dat"
    f, err := os.Open(fileName)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    fmt.Println("--- PHASE 1: STORING 1GB FILE ---")
    // Note: Since your Store now handles the disk write, 
    // we pass the file reader directly.
    if err := s3.Store("huge_file_001", f); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Done storing and broadcasting.")

    fmt.Println("--- PHASE 2: LOCAL DELETE ---")
    s3.store.Delete(s3.ID, "huge_file_001")

    fmt.Println("--- PHASE 3: NETWORK FETCH (The Stress Test) ---")
    startTime := time.Now()
    
    r, err := s3.Get("huge_file_001")
    if err != nil {
        log.Fatal(err)
    }

    // Use io.Copy to io.Discard to simulate reading the whole file 
    // without using any memory or printing to console.
    n, err := io.Copy(io.Discard, r)
    if err != nil {
        log.Fatal(err)
    }

    duration := time.Since(startTime)
    fmt.Printf("SUCCESS! Fetched %d bytes in %v\n", n, duration)

 
}
