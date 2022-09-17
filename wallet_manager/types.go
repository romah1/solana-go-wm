package wallet_manager

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"time"
)

type WalletManager struct {
	Context                context.Context
	Client                 *rpc.Client
	Commitment             rpc.CommitmentType
	ConfirmationStatusType rpc.ConfirmationStatusType
	ConfirmationTimeout    time.Duration
	ConfirmationDelay      time.Duration
	SkipPreflight          bool
}

type SendLamportsInstructionParams struct {
	From     solana.PrivateKey
	To       solana.PublicKey
	Lamports uint64
}

type SendSolInstructionParams struct {
	From solana.PrivateKey
	To   solana.PublicKey
	Sol  float64
}

func (params *SendSolInstructionParams) toLamports() SendLamportsInstructionParams {
	return SendLamportsInstructionParams{
		params.From,
		params.To,
		uint64(params.Sol * float64(solana.LAMPORTS_PER_SOL)),
	}
}
