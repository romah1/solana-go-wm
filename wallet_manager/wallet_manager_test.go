package wallet_manager

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
	"testing"
	"time"
)

var ctx = context.TODO()
var client = rpc.New(rpc.DevNet.RPC)
var commitment = rpc.CommitmentConfirmed
var confirmationCommitment = rpc.ConfirmationStatusConfirmed
var confirmationTimeout = time.Duration(5) * time.Minute
var confirmationDelay = time.Duration(5) * time.Second
var wm = NewWalletManagerWithOpts(ctx, client, commitment, confirmationCommitment, confirmationTimeout, confirmationDelay, false)

func TestWalletManager_SendLamports(t *testing.T) {
	var cntReceivers uint64 = 5
	lamportsPerReceiver := uint64(0.001 * float64(solana.LAMPORTS_PER_SOL))
	from := solana.NewWallet()
	_, err := airdrop(from.PublicKey(), (cntReceivers+1)*lamportsPerReceiver)
	if err != nil {
		t.Fatalf("failed to request airdrop: %s", err.Error())
	}
	var params []SendLamportsInstructionParams
	for i := uint64(0); i < cntReceivers; i++ {
		params = append(params, SendLamportsInstructionParams{
			From:     from.PrivateKey,
			To:       solana.NewWallet().PublicKey(),
			Lamports: lamportsPerReceiver,
		})
	}

	sig, err := wm.SendLamportsTransaction(from.PublicKey(), params)
	if err != nil {
		t.Fatalf("failed to spread lamports. err: %s", err.Error())
	}
	for _, param := range params {
		receiver := param.To
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

func TestWalletManager_SendAllSol(t *testing.T) {
	lamports := uint64(0.001 * float64(solana.LAMPORTS_PER_SOL))
	from := solana.NewWallet()
	_, err := airdrop(from.PublicKey(), lamports)
	if err != nil {
		t.Fatal(err)
	}
	to := solana.NewWallet()
	sig, err := wm.SendAllSol(from.PrivateKey, to.PublicKey())
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.GetBalance(ctx, from.PublicKey(), commitment)
	if err != nil {
		t.Fatalf("failed to check %s balance. err: %s", from.PublicKey().String(), err.Error())
	}
	if result.Value != 0 {
		t.Fatalf("sender %s balance is not zero", from.PublicKey().String())
	}
	t.Log(sig)
}

func airdrop(receiver solana.PublicKey, lamports uint64) (solana.Signature, error) {
	airDropSig, err := client.RequestAirdrop(ctx, receiver, lamports, commitment)
	if err != nil {
		return solana.Signature{}, errors.Errorf("failed to request airdrop: %s", err.Error())
	}
	awaitedSig, err := wm.awaitSignaturesConfirmation([]solana.Signature{airDropSig})
	if err != nil {
		return solana.Signature{}, errors.Errorf("failed to confirm airdrop: %s", err.Error())
	}
	return awaitedSig, err
}
