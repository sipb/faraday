package remote

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
	"util/testkeyutil"
)

func CreateContextPair(t *testing.T, handlerA RequestHandler, handlerB RequestHandler) (LocalContext, LocalContext) {
	akey, acert := testkeyutil.GenerateTLSKeypairForTests(t, "cert-for-a", []string{"localhost"}, nil, nil, nil)
	certpool_for_a := x509.NewCertPool()
	a := LocalContext{
		Timeout:   time.Millisecond * 100,
		Handler:   handlerA,
		LocalCert: tls.Certificate{Certificate: [][]byte{acert.Raw}, PrivateKey: akey},
		RootCA:    certpool_for_a,
	}

	bkey, bcert := testkeyutil.GenerateTLSKeypairForTests(t, "cert-for-b", []string{"localhost"}, nil, nil, nil)
	certpool_for_b := x509.NewCertPool()
	b := LocalContext{
		Timeout:   time.Millisecond * 100,
		Handler:   handlerB,
		LocalCert: tls.Certificate{Certificate: [][]byte{bcert.Raw}, PrivateKey: bkey},
		RootCA:    certpool_for_b,
	}

	certpool_for_a.AddCert(bcert)
	certpool_for_b.AddCert(acert)

	return a, b
}

type SendStruct struct {
	ABC int
	DEF string
}

type RecvStruct struct {
	X123 int
	X456 string
}

func TestEndToEnd(t *testing.T) {
	af := func(remote_principal string, parse func(interface{}) error) (interface{}, error) {
		ss := &SendStruct{}
		if err := parse(ss); err != nil {
			return nil, err
		}
		return &RecvStruct{X123: -ss.ABC, X456: ss.DEF}, nil
	}
	bf := func(remote_principal string, parse func(interface{}) error) (interface{}, error) {
		t.Error("should not be here")
		return nil, errors.New("should not be here")
	}
	a, b := CreateContextPair(t, af, bf)
	stop, cherr, err := a.StartServe("localhost:1836")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		stop()
		err := <-cherr
		if err != nil && err != http.ErrServerClosed {
			t.Error(err)
		}
	}()
	conn := b.ConnectRemote("cert-for-a", "localhost:1836")

	rs := &RecvStruct{}
	if err := conn.Send(SendStruct{ABC: 6674, DEF: "gravitational-singularity"}, rs); err != nil {
		t.Fatal(err)
	}
	if rs.X123 != -6674 || rs.X456 != "gravitational-singularity" {
		t.Error("mismatched response from server:", rs)
	}
}

func LaunchProxy(t *testing.T, bind string, direct_to string) (func() int, error) {
	listener, err := net.Listen("tcp", bind)
	if err != nil {
		return nil, err
	}
	done := make(chan struct{})
	finished := make(chan int)
	go func() {
		count := 0
		defer func() {
			finished <- count
		}()
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err == nil {
				count++
				go func(conn net.Conn) {
					defer conn.Close()
					nconn, err := net.Dial("tcp", direct_to)
					if err != nil {
						t.Error(err)
					}
					defer nconn.Close()
					go func() {
						_, err := io.Copy(conn, nconn)
						if err != nil {
							t.Error(err)
						}
					}()
					_, err = io.Copy(nconn, conn)
					if err != nil {
						t.Error(err)
					}
				}(conn)
			} else {
				select {
				case <-done:
				default:
					t.Error(err)
				}
				break
			}
		}
	}()
	return func() int {
		err := listener.Close()
		if err != nil {
			t.Error(err)
		}
		done <- struct{}{}
		return <-finished
	}, nil
}

func TestConnectionReuse(t *testing.T) {
	af := func(remote_principal string, parse func(interface{}) error) (interface{}, error) {
		ss := &SendStruct{}
		if err := parse(ss); err != nil {
			return nil, err
		}
		return &RecvStruct{X123: -ss.ABC, X456: ss.DEF}, nil
	}
	bf := func(remote_principal string, parse func(interface{}) error) (interface{}, error) {
		t.Error("should not be here")
		return nil, errors.New("should not be here")
	}
	a, b := CreateContextPair(t, af, bf)
	stop, cherr, err := a.StartServe("localhost:1836")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		stop()
		err := <-cherr
		if err != nil && err != http.ErrServerClosed {
			t.Error(err)
		}
	}()
	stop2, err := LaunchProxy(t, "localhost:1837", "localhost:1836")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		count := stop2()
		if count != 1 {
			t.Error("involved", count, " tcp connections, rather than one TCP connection")
		}
	}()
	conn := b.ConnectRemote("cert-for-a", "localhost:1837")

	// try a hundred requests. they should all go over the same TCP connection.
	for i := 1; i <= 100; i++ {
		rs := &RecvStruct{}
		if err := conn.Send(SendStruct{ABC: 666 + i, DEF: "brimstone"}, rs); err != nil {
			t.Fatal(err)
		}
		if rs.X123 != -666-i || rs.X456 != "brimstone" {
			t.Error("mismatched response from server:", rs)
		}
	}
}

func TestNeedsCorrectAuth(t *testing.T) {
	af := func(remote_principal string, parse func(interface{}) error) (interface{}, error) {
		t.Error("should not be here")
		return nil, errors.New("should not be here")
	}
	bf := func(remote_principal string, parse func(interface{}) error) (interface{}, error) {
		t.Error("should not be here")
		return nil, errors.New("should not be here")
	}
	a, b := CreateContextPair(t, af, bf)
	a.RootCA = b.RootCA // make it authenticate with the wrong CA
	stop, cherr, err := a.StartServe("localhost:1836")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		stop()
		err := <-cherr
		if err != nil && err != http.ErrServerClosed {
			t.Error(err)
		}
	}()
	conn := b.ConnectRemote("cert-for-a", "localhost:1836")

	rs := &RecvStruct{}
	err = conn.Send(SendStruct{ABC: 6674, DEF: "gravitational-singularity"}, rs)
	if err == nil {
		t.Error("should have been an error")
	} else if !strings.Contains(err.Error(), "remote error: tls: bad certificate") {
		t.Error("wrong error", err)
	}
}

func TestNeedsAnyAuth(t *testing.T) {
	af := func(remote_principal string, parse func(interface{}) error) (interface{}, error) {
		t.Error("should not be here")
		return nil, errors.New("should not be here")
	}
	bf := func(remote_principal string, parse func(interface{}) error) (interface{}, error) {
		t.Error("should not be here")
		return nil, errors.New("should not be here")
	}
	a, b := CreateContextPair(t, af, bf)
	stop, cherr, err := a.StartServe("localhost:1836")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		stop()
		err := <-cherr
		if err != nil && err != http.ErrServerClosed {
			t.Error(err)
		}
	}()
	conn := b.ConnectRemote("cert-for-a", "localhost:1836")
	conn.client.Transport.(*http.Transport).TLSClientConfig.Certificates = nil

	rs := &RecvStruct{}
	err = conn.Send(SendStruct{ABC: 6674, DEF: "gravitational-singularity"}, rs)
	if err == nil {
		t.Error("should have been an error")
	} else if !strings.Contains(err.Error(), "remote error: tls: bad certificate") {
		t.Error("wrong error", err)
	}
}
