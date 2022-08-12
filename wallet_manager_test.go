package solana_go_wm

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"testing"
	"time"
)

var ctx = context.TODO()
var client = rpc.New(rpc.DevNet.RPC)
var commitment = rpc.CommitmentConfirmed
var confirmationCommitment = rpc.ConfirmationStatusConfirmed
var confirmationTimeout = time.Duration(5) * time.Minute
var confirmationDelay = time.Duration(5) * time.Second
var wm = NewWalletManager(ctx, client, commitment, confirmationCommitment, confirmationTimeout, confirmationDelay, false)

func TestSendLamports(t *testing.T) {
	var cntReceivers uint64 = 5
	lamportsPerReceiver := uint64(0.001 * float64(solana.LAMPORTS_PER_SOL))
	from := solana.NewWallet()
	airDropSig, err := client.RequestAirdrop(ctx, from.PublicKey(), (cntReceivers+1)*lamportsPerReceiver, commitment)
	if err != nil {
		t.Fatalf("failed to request airdrop: %s", err.Error())
	}
	awaitedSig, err := wm.awaitSignaturesConfirmation([]solana.Signature{airDropSig})
	if err != nil {
		t.Fatalf("failed to confirm airdrop: %s", err.Error())
	}
	if airDropSig != awaitedSig {
		t.Fatalf("airdrop sig %s != awaited sig %s", airDropSig, awaitedSig)
	}
	var receivers []solana.PublicKey
	for i := uint64(0); i < cntReceivers; i++ {
		receivers = append(receivers, solana.NewWallet().PublicKey())
	}

	sig, err := wm.SpreadLamports(from.PrivateKey, receivers, lamportsPerReceiver)
	if err != nil {
		t.Fatalf("failed to spread lamports. err: %s", err.Error())
	}
	for _, receiver := range receivers {
		result, err := client.GetBalance(ctx, receiver, commitment)
		if err != nil {
			t.Fatalf("failed to get balance of %s", receiver.String())
		}
		if result.Value != lamportsPerReceiver {
			t.Fatalf("account %s balance is %d != %d", receiver.String(), result.Value, lamportsPerReceiver)
		}
	}
	t.Log(sig.String())
}
