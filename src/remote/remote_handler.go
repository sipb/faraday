// Package remote provides an abstraction for sending JSON objects over HTTPS within a network.
// To use it, construct an instance of LocalContext and call ConnectRemote or StartServe.
package remote

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

// A RequestHandler is a function that can be used to handle incoming requests to a serving LocalContext. Data is
// received in the form of a JSON object, and the RequestHandler retrieves this by calling parse on a prepared object,
// which decodes the data into that object via json.Unmarshal. The result will be encoded with json.Marshal and
// transmitted back to the requesting system. remote_principal is the CommonName on the TLS cert used by the remote
// system to perform authentication. The data transferred to and from this function is both authenticated and
// confidential, by the security properties of TLS.
type RequestHandler func(remote_principal string, parse func(interface{}) error) (interface{}, error)

// A Remote is a representation of a connection between the local system and a remote system. The existence of
// a Remote does not imply that a TCP or HTTPS connection has actually been established.
// Data can be transferred over a Remote by calling the Send() method.
type Remote struct {
	manager    *LocalContext
	client     http.Client
	expectedCN string
	addr       string
}

// A LocalContext is a representation of a local endpoint that can either handle requests from other systems, or
// generate new requests to send to other systems. Use ConnectRemote() to connect to another system, in preparation for
// sending data, and use StartServe() to start handling requests from other systems.
// You should make sure to fill out all of the fields on a LocalContext, except perhaps Handler.
type LocalContext struct {
	// The pool of certificate authorities that this system accepts certificates from, both when connecting to other
	// systems and when receiving requests from other systems.
	RootCA *x509.CertPool
	// The certificate used to authenticate this local system to other systems, both when connecting to other systems
	// and when receiving requests from other systems.
	LocalCert tls.Certificate
	// The handler used when a request is received from another system.
	Handler RequestHandler
	// The timeout used for all requests, in and out of the remote.
	Timeout time.Duration
}

// ConnectRemote prepares to connect to a particular remote system. 'remoteName' is the expected principal of the remote
// system, as specified in the CommonName field of its TLS certificate, and 'addr' is the address (including port) at
// which the remote system should be reachable. It returns a Remote object, which can be used to transfer
// individual requests. Where possible, open HTTPS connections to the server will be reused, to avoid re-negotiating TLS
// connections.
//
// Warning: successful return from this function does not imply that a connection has established; this is deferred
// until the first actual message sent.
func (manager *LocalContext) ConnectRemote(remoteName string, addr string) Remote {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      manager.RootCA,
				Certificates: []tls.Certificate{manager.LocalCert},
				MinVersion:   tls.VersionTLS12,
			},
			DisableCompression: true,
		},
		Timeout: manager.Timeout,
	}
	return Remote{manager, client, remoteName, addr}
}

func (manager *LocalContext) verifyTLS(tls *tls.ConnectionState, isclient bool) (string, error) {
	if tls == nil || len(tls.VerifiedChains) == 0 || len(tls.VerifiedChains[0]) == 0 {
		return "", errors.New("no certificate")
	} else {
		firstCert := tls.VerifiedChains[0][0]
		// might be duplicate work, but this guarantees it's correct
		var auth x509.ExtKeyUsage
		if isclient {
			auth = x509.ExtKeyUsageClientAuth
		} else {
			auth = x509.ExtKeyUsageServerAuth
		}
		chains, err := firstCert.Verify(x509.VerifyOptions{
			Roots:     manager.RootCA,
			KeyUsages: []x509.ExtKeyUsage{auth},
		})
		if err != nil || len(chains) == 0 {
			return "", errors.New("no valid certificate")
		} else {
			return firstCert.Subject.CommonName, nil
		}
	}
}

// Send transmits an individual request across the network to this remote system. The specified message is encoded as a
// JSON object with json.Marshal. The result of the request is decoded (with json.Unmarshal) into the result parameter.
// The request will be handled by the RequestHandler on the remote end.
// The data transmitted here is both authenticated and confidential, through the security properties of TLS.
func (conn *Remote) Send(message interface{}, result interface{}) error {
	reqbody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("while marshalling json for request: %s", err.Error())
	}
	response, err := conn.client.Post("https://"+conn.addr+"/faraday", "application/json", bytes.NewReader(reqbody))
	if err != nil {
		return fmt.Errorf("while processing request: %s", err.Error())
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}
	principal, err := conn.manager.verifyTLS(response.TLS, false)
	if err != nil {
		return err
	}
	if principal != conn.expectedCN {
		return fmt.Errorf("mismatched common name while receiving response")
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("while receiving response: %s", err.Error())
	}
	err = response.Body.Close()
	if err != nil {
		return fmt.Errorf("while closing connection: %s", err.Error())
	}
	err = json.Unmarshal(body, result)
	if err != nil {
		return fmt.Errorf("while unmarshalling json from response: %s", err.Error())
	}
	return nil
}

// StartServe launches a local HTTPS server that receives requests for this system, as sent via Remote.Send(). As it
// receives requests, it calls into the RequestHandler registered on the manager, passing it the ability to decode the
// incoming JSON request, and encoding the result. It returns three results: if it fails to initialize the server, it
// will return an error in the third result. Otherwise, the first result is a function that can be called to stop the
// server, and the second result is a channel from which the exit error can be read, if the HTTP server ever stops
// serving, such as by calling the stop function.
func (manager *LocalContext) StartServe(addr string) (func(), chan error, error) {
	// recommended addr: ":1836" (the year the faraday cage was invented)
	server := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path != "/faraday" {
				http.Error(writer, "not found", 404)
				return
			}
			principal, err := manager.verifyTLS(request.TLS, true)
			if err != nil {
				http.Error(writer, "no certificate", 403)
				return
			}
			data, err := ioutil.ReadAll(request.Body)
			if err != nil {
				http.Error(writer, "failed to read data", 400)
				return
			}
			result, err := manager.Handler(principal, func(output interface{}) error {
				return json.Unmarshal(data, output)
			})
			if err != nil {
				log.Println("Failed:", err, "during request from", principal)
				http.Error(writer, "request failed", 500)
				return
			}
			to_write, err := json.Marshal(result)
			if err != nil {
				http.Error(writer, "failed to write data", 500)
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			writer.Write(to_write) // TODO: do I need to handle errors from this? does it matter?
		}),
		TLSConfig: &tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    manager.RootCA,
			Certificates: []tls.Certificate{manager.LocalCert},
			MinVersion:   tls.VersionTLS12,
			NextProtos:   []string{"http/1.1", "h2"},
		},
		ReadTimeout:  manager.Timeout,
		WriteTimeout: manager.Timeout,
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return nil, nil, err
	}

	cherr := make(chan error)

	go func() {
		tlsListener := tls.NewListener(ln, server.TLSConfig)
		cherr <- server.Serve(tlsListener)
	}()

	return func() { server.Shutdown(context.Background()) }, cherr, nil
}
