package network

import (
	"bytes"
	"github.com/LemoFoundationLtd/lemochain-go/chain/deputynode"
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/log"
	"github.com/LemoFoundationLtd/lemochain-go/common/subscribe"
	"github.com/LemoFoundationLtd/lemochain-go/network/p2p"
	"net"
	"time"
)

const (
	ReconnectNode = "reconnectNode"
)

type DialManager struct {
	coreNodeID       *p2p.NodeID
	coreNodeEndpoint string
}

func NewDialManager(coreNodeID *p2p.NodeID, coreNodeEndpoint string) *DialManager {
	return &DialManager{
		coreNodeID:       coreNodeID,
		coreNodeEndpoint: coreNodeEndpoint,
	}
}

// Dial run dial
func (dm *DialManager) Dial() {
	// dial
	conn, err := net.DialTimeout("tcp", dm.coreNodeEndpoint, 3*time.Second)
	if err != nil {
		log.Warnf("dial node error: %s", err.Error())
		subscribe.Send(ReconnectNode, struct{}{})
		return
	}
	// handle connection
	if err = dm.handleConn(conn); err != nil {
		subscribe.Send(ReconnectNode, struct{}{})
		return
	}
}

// handleConn handle the connection
func (dm *DialManager) handleConn(fd net.Conn) error {
	p := p2p.NewPeer(fd)
	if err := p.DoHandshake(deputynode.GetSelfNodeKey(), dm.coreNodeID); err != nil {
		if err = fd.Close(); err != nil {
			log.Errorf("close connection failed: %v", err)
		}
		return err
	}
	// is self
	if bytes.Compare(p.RNodeID()[:], deputynode.GetSelfNodeID()) == 0 {
		if err := fd.Close(); err != nil {
			log.Errorf("close connections failed", err)
		} else {
			log.Error("can't connect self")
		}
		return p2p.ErrConnectSelf
	}
	go dm.runPeer(p)
	subscribe.Send(subscribe.AddNewPeer, p)
	return nil
}

// runPeer run the connected peer
func (dm *DialManager) runPeer(p p2p.IPeer) {
	if err := p.Run(); err != nil { // block this
		log.Debugf("runPeer error: %v", err)
	}
	subscribe.Send(ReconnectNode, struct{}{})
	log.Debugf("peer Run finished: %s", common.ToHex(p.RNodeID()[:8]))
}
