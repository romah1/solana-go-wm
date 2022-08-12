package solana_go_wm

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
	"time"
)

func sendAndConfirmInstructions(
	ctx context.Context,
	client *rpc.Client,
	feePayer solana.PublicKey,
	instructions []solana.Instruction,
	signers []solana.PrivateKey,
) (solana.Signature, error) {
	return sendAndConfirmInstructionsWithOpts(
		ctx,
		client,
		feePayer,
		instructions,
		signers,
		false,
		commitment,
		time.Duration(30)*time.Second,
		time.Duration(1)*time.Second,
	)
}

func sendAndConfirmInstructionsWithOpts(
	ctx context.Context,
	client *rpc.Client,
	feePayer solana.PublicKey,
	instructions []solana.Instruction,
	signers []solana.PrivateKey,
	skipPreflight bool,
	commitment rpc.CommitmentType,
	confirmationTimeout time.Duration,
	confirmationDelay time.Duration,
) (solana.Signature, error) {
	recent, err := client.GetRecentBlockhash(ctx, commitment)
	if err != nil {
		return solana.Signature{}, err
	}
	txBuilder := solana.NewTransactionBuilder().
		SetRecentBlockHash(recent.Value.Blockhash).
		SetFeePayer(feePayer)
	for _, instruction := range instructions {
		txBuilder.AddInstruction(instruction)
	}
	tx, err := txBuilder.Build()
	if err != nil {
		return solana.Signature{}, err
	}
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		for _, candidate := range signers {
			if candidate.PublicKey().Equals(key) {
				return &candidate
			}
		}
		return nil
	})
	if err != nil {
		return solana.Signature{}, err
	}
	return sendAndConfirmTransaction(ctx, client, tx, skipPreflight, commitment, confirmationTimeout, confirmationDelay)
}

func sendAndConfirmTransaction(
	ctx context.Context,
	client *rpc.Client,
	tx *solana.Transaction,
	skipPreflight bool,
	commitment rpc.CommitmentType,
	confirmationTimeout time.Duration,
	confirmationDelay time.Duration,
) (solana.Signature, error) {
	sig, err := client.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
		SkipPreflight:       skipPreflight,
		PreflightCommitment: commitment,
	})
	if err != nil {
		return solana.Signature{}, err
	}
	return awaitSignaturesConfirmation(client, []solana.Signature{sig}, commitment, confirmationTimeout, confirmationDelay)
}

func awaitSignaturesConfirmation(
	client *rpc.Client,
	signatures []solana.Signature,
	commitment rpc.CommitmentType,
	timeout time.Duration,
	delay time.Duration,
) (solana.Signature, error) {
	if len(signatures) == 0 {
		return solana.Signature{}, errors.New("signatures array is empty")
	}
	after := time.After(timeout)
	ticker := time.NewTicker(delay)

	for {
		select {
		case <-ticker.C:
			for _, sig := range signatures {
				result, err := client.GetTransaction(context.TODO(), sig, &rpc.GetTransactionOpts{
					Commitment: commitment,
				})
				if err == nil && result.Meta.Err == nil {
					spew.Dump("Success sig: " + sig.String())
					return sig, nil
				}
			}
		case <-after:
			return solana.Signature{}, errors.New("timeout")
		}
	}
}
