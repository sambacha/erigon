package commands

import (
	"fmt"
	"testing"

	"github.com/ledgerwatch/erigon-lib/gointerfaces/remote"
	"github.com/ledgerwatch/erigon-lib/gointerfaces/sentry"
	"github.com/ledgerwatch/erigon/cmd/rpcdaemon/rpcdaemontest"
	"github.com/ledgerwatch/erigon/cmd/rpcdaemon/rpcservices"
	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/core"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/eth/protocols/eth"
	"github.com/ledgerwatch/erigon/rlp"
	"github.com/ledgerwatch/erigon/turbo/rpchelper"
	"github.com/ledgerwatch/erigon/turbo/snapshotsync"
	"github.com/ledgerwatch/erigon/turbo/stages"
	"github.com/stretchr/testify/require"
)

func TestEthSubscribe(t *testing.T) {
	m, require := stages.Mock(t), require.New(t)
	chain, err := core.GenerateChain(m.ChainConfig, m.Genesis, m.Engine, m.DB, 21, func(i int, b *core.BlockGen) {
		b.SetCoinbase(common.Address{1})
	}, false /* intermediateHashes */)
	require.NoError(err)

	b, err := rlp.EncodeToBytes(&eth.BlockHeadersPacket66{
		RequestId:          1,
		BlockHeadersPacket: chain.Headers,
	})
	require.NoError(err)

	m.ReceiveWg.Add(1)
	for _, err = range m.Send(&sentry.InboundMessage{Id: sentry.MessageId_BLOCK_HEADERS_66, Data: b, PeerId: m.PeerId}) {
		require.NoError(err)
	}
	m.ReceiveWg.Wait() // Wait for all messages to be processed before we proceeed

	ctx, conn := rpcdaemontest.CreateTestGrpcConn(t, m)
	backend := rpcservices.NewRemoteBackend(remote.NewETHBACKENDClient(conn), m.DB, snapshotsync.NewBlockReader())
	ff := rpchelper.New(ctx, backend, nil, nil, func() {})

	newHeads := make(chan *types.Header)
	id := ff.SubscribeNewHeads(newHeads)
	defer ff.UnsubscribeHeads(id)

	initialCycle := true
	highestSeenHeader := chain.TopBlock.NumberU64()
	if _, err := stages.StageLoopStep(m.Ctx, m.DB, m.Sync, highestSeenHeader, m.Notifications, initialCycle, m.UpdateHead, nil); err != nil {
		t.Fatal(err)
	}

	for i := uint64(1); i <= highestSeenHeader; i++ {
		header := <-newHeads
		fmt.Printf("Got header %d\n", header.Number.Uint64())
		require.Equal(i, header.Number.Uint64())
	}
}
