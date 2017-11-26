package main

import (
	"common"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"farad/history"
	"farad/membership"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"remote"
	"sync"
	"time"
	"util/timeutil"
	"util/wraputil"
)

type State struct {
	members *membership.MemberContext
	hist    *history.History
	lock    sync.Mutex
}

func GenServerId() (string, error) {
	server_id := make([]byte, 16)
	_, err := rand.Read(server_id)
	if err != nil {
		return "", fmt.Errorf("while generating server ID: %s", err.Error())
	}
	return hex.EncodeToString(server_id), nil
}

func FaradMain(authority *x509.Certificate, cert tls.Certificate) error {
	state := State{
		members: membership.NewMemberContext(time.Second * 2), // expire after two seconds without contact
		hist:    history.NewHistory(500),
	}

	server_id, err := GenServerId()
	if err != nil {
		return err
	}

	pool := x509.NewCertPool()
	pool.AddCert(authority)
	server := remote.LocalContext{
		RootCA:    pool,
		Timeout:   time.Millisecond * 100,
		LocalCert: cert,
		Handler: func(remote_principal string, parse func(interface{}) error) (interface{}, error) {
			req := &common.FaradRequest{}
			if err := parse(req); err != nil {
				return nil, err
			}
			if req.Version != common.FARADAY_PROTOCOL_VERSION {
				return nil, fmt.Errorf("wrong faraday version: %d instead of %d", req.Version, common.FARADAY_PROTOCOL_VERSION)
			}
			if req.ServerInstance != server_id {
				// this must be a new server (or the wrong server...?) -- so we should send everything
				req.Cursor = 0
			}
			state.lock.Lock()
			defer state.lock.Unlock()
			did_revision_occur, err := state.members.UpdatePing(remote_principal, req.Key)
			if err != nil {
				return nil, err
			}
			if did_revision_occur {
				state.hist.AddUpdate(remote_principal)
			}
			has_all, changes, now := state.hist.Since(req.Cursor)
			response := &common.FaradResponse{
				Cursor:         now,
				ServerInstance: server_id,
			}
			if has_all {
				if req.IncludeMember != "" {
					changes = append(changes, req.IncludeMember)
				}
				response.CurrentCluster = state.members.Subshot(changes)
			} else {
				response.CurrentCluster = state.members.Snapshot()
			}
			return response, nil
		},
	}
	stop, cherr, err := server.StartServe(":1836")
	defer stop()
	if err != nil {
		return err
	}

	return <-cherr
}

func main() {
	if len(os.Args) != 4 {
		log.Fatalln("Usage: farad <ca-path> <cert-path> <key-path>")
	}
	ca_data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalln("Could not read CA:", err)
	}
	ca, err := wraputil.LoadX509CertFromPEM(ca_data)
	if err != nil {
		log.Fatalln("Could not parse CA:", err)
	}
	tcert, err := tls.LoadX509KeyPair(os.Args[2], os.Args[3])
	if err != nil {
		log.Fatalln("Could not load cert:", err)
	}
	err = FaradMain(ca, tcert)
	if err != nil {
		log.Fatalln("farad failed:", err)
	}
}
