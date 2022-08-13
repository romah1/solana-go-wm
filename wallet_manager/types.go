package wallet_manager

import (
	"context"
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
