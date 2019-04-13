package network

import (
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"github.com/LemoFoundationLtd/lemochain-core/common/rlp"
	coreNetwork "github.com/LemoFoundationLtd/lemochain-core/network"
	"github.com/LemoFoundationLtd/lemochain-core/network/p2p"
	"sync"
	"time"
)

const (
	DurShort = 3 * time.Second
	DurLong  = 10 * time.Second
)

// ProtocolHandshake protocol handshake
type ProtocolHandshake struct {
	ChainID      uint16
	GenesisHash  common.Hash
	NodeVersion  uint32
	LatestStatus LatestStatus
}

// Bytes object to bytes
func (phs *ProtocolHandshake) Bytes() []byte {
	buf, err := rlp.EncodeToBytes(phs)
	if err != nil {
		return nil
	}
	return buf
}

// LatestStatus latest peer's status
type LatestStatus struct {
	CurHeight uint32
	CurHash   common.Hash

	StaHeight uint32
	StaHash   common.Hash
}

type peer struct {
	conn p2p.IPeer

	lstStatus LatestStatus

	firstSyncHeight uint32 // first sync height when handlePeer()

	lock sync.RWMutex
}

// newPeer new peer instance
func newPeer(p p2p.IPeer) *peer {
	return &peer{
		conn: p,
	}
}

// GetFirstSyncHeight
func (p *peer) GetFirstSyncHeight() uint32 {
	p.lock.Lock()
	defer p.lock.Unlock()
	height := p.firstSyncHeight
	return height
}

// SetFirstSyncHeight
func (p *peer) SetFirstSyncHeight(height uint32) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.firstSyncHeight = height
}

// NormalClose
func (p *peer) NormalClose() {
	p.conn.Close()
	p.conn.SetStatus(p2p.StatusNormal)
}

// ManualClose
func (p *peer) ManualClose() {
	p.conn.Close()
	p.conn.SetStatus(p2p.StatusManualDisconnect)
}

// FailedHandshakeClose
func (p *peer) FailedHandshakeClose() {
	p.conn.Close()
	p.conn.SetStatus(p2p.StatusFailedHandshake)
}

// RcvBadDataClose
func (p *peer) RcvBadDataClose() {
	p.conn.Close()
	p.conn.SetStatus(p2p.StatusBadData)
}

// HardForkClose
func (p *peer) HardForkClose() {
	p.conn.Close()
	p.conn.SetStatus(p2p.StatusHardFork)
}

// RequestBlocks request blocks from remote
func (p *peer) RequestBlocks(from, to uint32) int {
	if from > to {
		log.Warnf("RequestBlocks: from: %d can't be larger than to:%d", from, to)
		return -1
	}
	msg := &coreNetwork.GetBlocksData{From: from, To: to}
	buf, err := rlp.EncodeToBytes(&msg)
	if err != nil {
		log.Warnf("RequestBlocks: rlp encode failed: %v", err)
		return -2
	}
	// log.Debugf("start request blocks: from: %d, to: %d", from, to)
	p.conn.SetWriteDeadline(DurShort)
	if err = p.conn.WriteMsg(coreNetwork.GetBlocksWithChangeLogMsg, buf); err != nil {
		log.Warnf("RequestBlocks: write message failed: %v", err)
		return -3
	}
	log.Debugf("Request Blocks height from %d to %d", from, to)
	return 0
}

// Handshake protocol handshake
func (p *peer) Handshake(content []byte) (*ProtocolHandshake, error) {
	// write to remote
	if err := p.conn.WriteMsg(coreNetwork.ProHandshakeMsg, content); err != nil {
		return nil, err
	}
	// read from remote
	msgCh := make(chan *p2p.Msg)
	go func() {
		if msg, err := p.conn.ReadMsg(); err == nil {
			msgCh <- msg
		} else {
			log.Warnf("when handshake but read message error: %s", err)
			msgCh <- nil
		}
	}()
	timeout := time.NewTimer(8 * time.Second)
	select {
	case <-timeout.C:
		return nil, coreNetwork.ErrReadTimeout
	case msg := <-msgCh:
		if msg == nil {
			return nil, coreNetwork.ErrReadMsg
		}
		var phs ProtocolHandshake
		if err := msg.Decode(&phs); err != nil {
			return nil, err
		}
		return &phs, nil
	}
}

// NodeID
func (p *peer) NodeID() *p2p.NodeID {
	return p.conn.RNodeID()
}

// ReadMsg read message from net stream
func (p *peer) ReadMsg() (*p2p.Msg, error) {
	return p.conn.ReadMsg()
}

// SendLstStatus send SyncFailednode's status to remote
func (p *peer) SendLstStatus(status *LatestStatus) int {
	buf, err := rlp.EncodeToBytes(status)
	if err != nil {
		log.Warnf("SendLstStatus: rlp encode failed: %v", err)
		return -1
	}
	p.conn.SetWriteDeadline(DurShort)
	if err = p.conn.WriteMsg(coreNetwork.LstStatusMsg, buf); err != nil {
		log.Warnf("SendLstStatus to peer: %s failed: %v", p.NodeID().String()[:16], err)
		p.conn.Close()
		return -2
	}
	return 0
}

// SendTxs send txs to remote
func (p *peer) SendTxs(txs types.Transactions) int {
	buf, err := rlp.EncodeToBytes(&txs)
	if err != nil {
		log.Warnf("SendTxs: rlp failed: %v", err)
		return -1
	}
	if err := p.conn.WriteMsg(coreNetwork.TxsMsg, buf); err != nil {
		log.Warnf("SendTxs to peer: %s failed. disconnect. %v", p.NodeID().String()[:16], err)
		p.conn.Close()
		return -2
	}
	return 0
}

// SendReqLatestStatus send request of latest status
func (p *peer) SendReqLatestStatus() int {
	msg := &coreNetwork.GetLatestStatus{Revert: uint32(0)}
	buf, err := rlp.EncodeToBytes(msg)
	if err != nil {
		log.Warnf("SendReqLatestStatus: rlp failed: %v", err)
		return -1
	}
	p.conn.SetWriteDeadline(DurShort)
	if err := p.conn.WriteMsg(coreNetwork.GetLstStatusMsg, buf); err != nil {
		log.Warnf("SendReqLatestStatus to peer: %s failed. disconnect. %v", p.NodeID().String()[:16], err)
		p.conn.Close()
		return -2
	}
	return 0
}

// LatestStatus return record of latest status
func (p *peer) LatestStatus() LatestStatus {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.lstStatus
}

// UpdateStatus update peer's latest status
func (p *peer) UpdateStatus(height uint32, hash common.Hash) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.lstStatus.StaHeight < height {
		p.lstStatus.StaHeight = height
		p.lstStatus.StaHash = hash
	}
}
