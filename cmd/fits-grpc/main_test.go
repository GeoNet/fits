package main

import (
	"crypto/tls"
	"database/sql"
	tk "github.com/GeoNet/fits/internal/credentials/token"
	"github.com/GeoNet/fits/internal/fits"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"log"
	"os"
	"testing"
)

var testServer *grpc.Server
var conn *grpc.ClientConn
var connNoCreds *grpc.ClientConn

// TestMain starts the testServer and connections to it.
// A self signed TLS cert is auto generated and verification
// is skipped on the client connections.
func TestMain(m *testing.M) {
	tokenWrite = "testwrite"

	var err error

	var fs fitsServer

	fs.db, err = sql.Open("postgres", "host=localhost connect_timeout=300 user=fits_w password=test dbname=fits sslmode=disable statement_timeout=600000")
	if err != nil {
		log.Fatalf("ERROR: problem with DB config: %s", err)
	}

	fs.db.SetMaxIdleConns(5)
	fs.db.SetMaxOpenConns(15)

	defer fs.close()

	testServer = grpc.NewServer(grpc.UnaryInterceptor(telemetry))

	fits.RegisterFitsServer(testServer, &fs)

	cert, err := selfSigned()
	if err != nil {
		log.Fatalf("failed to generate self signed TLS cert: %v", err)
	}

	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	lis, err := tls.Listen("tcp", ":8443", &config)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Print("starting test server")
	go testServer.Serve(lis)

	conn, err = grpc.Dial("localhost:8443",
		grpc.WithPerRPCCredentials(tk.New("testwrite")),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{ServerName: "", InsecureSkipVerify: true})))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	connNoCreds, err = grpc.Dial("localhost:8443",
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{ServerName: "", InsecureSkipVerify: true})))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	code := m.Run()

	conn.Close()
	connNoCreds.Close()
	testServer.Stop()

	os.Exit(code)
}

func TestWrite(t *testing.T) {
	err := write(metadata.NewIncomingContext(context.Background(), metadata.Pairs()))
	if grpc.Code(err) != codes.Unauthenticated {
		t.Error("should get unuathenicated error with no token set")
	}

	err = write(metadata.NewIncomingContext(context.Background(), metadata.Pairs("token", "wrong")))
	if grpc.Code(err) != codes.Unauthenticated {
		t.Error("should get unuathenicated with wrong token set")
	}

	tokenWrite = ""

	err = write(metadata.NewIncomingContext(context.Background(), metadata.Pairs("token", "testwrite")))
	if grpc.Code(err) != codes.Unauthenticated {
		t.Error("should get unuathenicated with empty server token set")
	}

	tokenWrite = "wrong"

	err = write(metadata.NewIncomingContext(context.Background(), metadata.Pairs("token", "testwrite")))
	if grpc.Code(err) != codes.Unauthenticated {
		t.Error("should get unuathenicated with wrong server token set")
	}

	// reset server token to the testing one
	tokenWrite = "testwrite"

	err = write(metadata.NewIncomingContext(context.Background(), metadata.Pairs("token", "testwrite")))
	if err != nil {
		t.Error("should get no error with the correct token set")
	}
}
