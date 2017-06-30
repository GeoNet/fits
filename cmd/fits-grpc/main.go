package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/GeoNet/fits/internal/fits"
	"github.com/GeoNet/mtr/mtrapp"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"math/big"
	"os"
	"time"
)

var tokenWrite = os.Getenv("TOKEN_WRITE")

func main() {
	switch "" {
	case tokenWrite:
		log.Fatal("empty write token, exiting")
	}

	var fs fitsServer

	err := fs.init()
	if err != nil {
		log.Fatal(err)
	}
	defer fs.close()

	cert, err := selfSigned()
	if err != nil {
		log.Fatalf("failed to generate self signed TLS cert: %v", err)
	}

	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	lis, err := tls.Listen("tcp", ":"+os.Getenv("PORT"), &config)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(telemetry))

	fits.RegisterFitsServer(s, &fs)

	log.Print("starting server")
	log.Fatal(s.Serve(lis))
}

func token(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return ""
	}

	t := md["token"]

	if len(t) != 1 {
		return ""
	}

	return t[0]
}

func write(ctx context.Context) error {
	if tokenWrite == "" {
		return grpc.Errorf(codes.Unauthenticated, "server write token empty")
	}

	switch token(ctx) {
	case tokenWrite:
		return nil
	default:
		return grpc.Errorf(codes.Unauthenticated, "valid write token required")
	}
}

// selfSigned generates a self signed TLS certificate.
func selfSigned() (tls.Certificate, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1337),
		Subject: pkix.Name{
			Organization: []string{"seflie"},
		},
		SignatureAlgorithm:    x509.SHA512WithRSA,
		PublicKeyAlgorithm:    x509.ECDSA,
		NotBefore:             time.Now().AddDate(-1, 0, 0),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		BasicConstraintsValid: true,
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	p, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return tls.Certificate{}, err
	}

	b, err := x509.CreateCertificate(rand.Reader, ca, ca, &p.PublicKey, p)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.Certificate{
		Certificate: [][]byte{b},
		PrivateKey:  p,
	}, nil
}

// TODO stream interceptor

// telemetry is a UnaryServerInterceptor.
func telemetry(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	i, err := handler(ctx, req)

	mtrapp.Requests.Inc()

	if s, ok := status.FromError(err); ok {
		switch s.Code() {
		case codes.OK:
			mtrapp.StatusOK.Inc()
		case codes.InvalidArgument:
			mtrapp.StatusBadRequest.Inc()
		case codes.Unauthenticated:
			mtrapp.StatusUnauthorized.Inc()
		case codes.NotFound:
			mtrapp.StatusNotFound.Inc()
		case codes.FailedPrecondition:
			log.Printf("%s %s", info.FullMethod, err.Error())
			mtrapp.StatusInternalServerError.Inc()
		case codes.Unavailable:
			log.Printf("%s %s", info.FullMethod, err.Error())
			mtrapp.StatusServiceUnavailable.Inc()
		}
	}

	return i, err
}
