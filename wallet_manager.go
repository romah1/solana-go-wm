package solana_go_wm

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
)

func SendSol(client *rpc.Client, from solana.PrivateKey, to solana.PublicKey, amountSol float64) (solana.Signature, error) {
	return SpreadSol(client, from, []solana.PublicKey{to}, amountSol)
}

func SendLamports(client *rpc.Client, from solana.PrivateKey, to solana.PublicKey, lamports uint64) (solana.Signature, error) {
	return SpreadLamports(client, from, []solana.PublicKey{to}, lamports)
}

func SpreadSol(client *rpc.Client, from solana.PrivateKey, toWallets []solana.PublicKey, sol float64) (solana.Signature, error) {
	return SpreadLamports(client, from, toWallets, uint64(sol*float64(solana.LAMPORTS_PER_SOL)))
}

func SpreadLamports(client *rpc.Client, from solana.PrivateKey, receivers []solana.PublicKey, lamports uint64) (solana.Signature, error) {
	var instructions []solana.Instruction
	for _, receiver := range receivers {
		instructions = append(instructions, makeTransferInstruction(from.PublicKey(), receiver, lamports))
	}
	return sendAndConfirmInstructions(
		context.TODO(),
		client,
		from.PublicKey(),
		instructions,
		[]solana.PrivateKey{from},
	)
}

func SendAllSol(client *rpc.Client, from solana.PrivateKey, to solana.PublicKey) (solana.Signature, error) {
	return CollectAllSol(client, []solana.PrivateKey{from}, to)
}

func CollectAllSol(client *rpc.Client, fromWallets []solana.PrivateKey, to solana.PublicKey) (solana.Signature, error) {
	if len(fromWallets) == 0 {
		return solana.Signature{}, errors.New("no wallets to send from")
	}
	feePayer := fromWallets[0]
	feeTx, err := makeTransferTransaction(client, feePayer, to, 0)
	if err != nil {
		return solana.Signature{}, errors.Errorf(
			"failed to make transfer transaction from %s to %s",
			feePayer.PublicKey().String(),
			to.String(),
		)
	}
	getFeeResult, err := client.GetFeeForMessage(context.TODO(), feeTx.Message.ToBase64(), commitment)
	if err != nil {
		return solana.Signature{}, errors.Errorf("failed to get fee for transaction %s", feeTx.String())
	}
	totalFee := *getFeeResult.Value * uint64(len(fromWallets))
	var instructions []solana.Instruction
	for i, from := range fromWallets {
		var fee uint64 = 0
		if i == 0 {
			fee = totalFee
		}
		balance, err := client.GetBalance(context.TODO(), from.PublicKey(), commitment)
		if err != nil {
			return solana.Signature{}, errors.Errorf("failed to get balance of %s", from.PublicKey().String())
		}
		instructions = append(instructions, makeTransferInstruction(from.PublicKey(), to, balance.Value-fee))
	}
	return sendAndConfirmInstructions(context.TODO(), client, feePayer.PublicKey(), instructions, fromWallets)
}

func makeTransferInstruction(from, to solana.PublicKey, lamports uint64) solana.Instruction {
	return system.NewTransferInstructionBuilder().
		SetFundingAccount(from).
		SetRecipientAccount(to).
		SetLamports(lamports).
		Build()
}

func makeTransferTransaction(client *rpc.Client, from solana.PrivateKey, to solana.PublicKey, lamports uint64) (*solana.Transaction, error) {
	instruction := makeTransferInstruction(from.PublicKey(), to, lamports)
	recent, err := client.GetRecentBlockhash(context.TODO(), commitment)
	if err != nil {
		return nil, err
	}
	tx, err := solana.NewTransactionBuilder().
		SetRecentBlockHash(recent.Value.Blockhash).
		AddInstruction(instruction).
		SetFeePayer(from.PublicKey()).
		Build()
	payload, err := tx.Message.MarshalBinary()
	if err != nil {
		return nil, err
	}
	sig, err := from.Sign(payload)
	if err != nil {
		return nil, err
	}
	tx.Signatures = append(tx.Signatures, sig)
	return tx, nil
}
