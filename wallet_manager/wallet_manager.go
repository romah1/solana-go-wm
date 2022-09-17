package wallet_manager

import (
	"context"
	"github.com/gagliardetto/solana-go"
	atok "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
	"time"
)

func NewWalletManager(client *rpc.Client) *WalletManager {
	return NewWalletManagerWithOpts(
		context.TODO(),
		client,
		rpc.CommitmentFinalized,
		rpc.ConfirmationStatusFinalized,
		time.Duration(30)*time.Second,
		time.Duration(5)*time.Second,
		false,
	)
}

func NewWalletManagerWithOpts(
	context context.Context,
	client *rpc.Client,
	commitment rpc.CommitmentType,
	confirmationStatusType rpc.ConfirmationStatusType,
	confirmationTimeout time.Duration,
	confirmationDelay time.Duration,
	skipPreflight bool,
) *WalletManager {
	return &WalletManager{
		Context:                context,
		Client:                 client,
		Commitment:             commitment,
		ConfirmationStatusType: confirmationStatusType,
		ConfirmationTimeout:    confirmationTimeout,
		ConfirmationDelay:      confirmationDelay,
		SkipPreflight:          skipPreflight,
	}
}

func (wm *WalletManager) SendSol(from solana.PrivateKey, to solana.PublicKey, amountSol float64) (solana.Signature, error) {
	return wm.SendSolTransaction(from, []SendSolInstructionParams{{from, to, amountSol}})
}

func (wm *WalletManager) SendLamports(from solana.PrivateKey, to solana.PublicKey, lamports uint64) (solana.Signature, error) {
	return wm.SendLamportsTransaction(from, []SendLamportsInstructionParams{{from, to, lamports}})
}

func (wm *WalletManager) SendSolTransaction(feePayer solana.PrivateKey, instructionsParams []SendSolInstructionParams) (solana.Signature, error) {
	var params []SendLamportsInstructionParams
	for _, solParams := range instructionsParams {
		params = append(params, solParams.toLamports())
	}
	return wm.SendLamportsTransaction(feePayer, params)
}

func (wm *WalletManager) SendLamportsTransaction(feePayer solana.PrivateKey, instructionsParams []SendLamportsInstructionParams) (solana.Signature, error) {
	var instructions []solana.Instruction
	var signers []solana.PrivateKey
	for _, params := range instructionsParams {
		instructions = append(instructions, makeTransferInstruction(params.From.PublicKey(), params.To, params.Lamports))
		signers = append(signers, params.From)
	}
	signers = appendSignerIfNotPresented(signers, feePayer)
	return wm.SendAndConfirmInstructions(
		feePayer.PublicKey(),
		instructions,
		signers,
	)
}

func (wm *WalletManager) SpreadLamports(from solana.PrivateKey, receivers []solana.PublicKey, lamports uint64) (solana.Signature, error) {
	var instructions []solana.Instruction
	for _, receiver := range receivers {
		instructions = append(instructions, makeTransferInstruction(from.PublicKey(), receiver, lamports))
	}
	return wm.SendAndConfirmInstructions(
		from.PublicKey(),
		instructions,
		[]solana.PrivateKey{from},
	)
}

func (wm *WalletManager) SendAllSol(from solana.PrivateKey, to solana.PublicKey) (solana.Signature, error) {
	return wm.CollectAllSol([]solana.PrivateKey{from}, to)
}

func (wm *WalletManager) CollectAllSol(fromWallets []solana.PrivateKey, to solana.PublicKey) (solana.Signature, error) {
	if len(fromWallets) == 0 {
		return solana.Signature{}, errors.New("no wallets to send from")
	}
	feePayer := fromWallets[0]
	feeTx, err := wm.makeTransferTransaction(feePayer, to, 0)
	if err != nil {
		return solana.Signature{}, errors.Errorf(
			"failed to make transfer transaction from %s to %s",
			feePayer.PublicKey().String(),
			to.String(),
		)
	}
	getFeeResult, err := wm.Client.GetFeeForMessage(context.TODO(), feeTx.Message.ToBase64(), wm.Commitment)
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
		balance, err := wm.Client.GetBalance(context.TODO(), from.PublicKey(), wm.Commitment)
		if err != nil {
			return solana.Signature{}, errors.Errorf("failed to get balance of %s", from.PublicKey().String())
		}
		instructions = append(instructions, makeTransferInstruction(from.PublicKey(), to, balance.Value-fee))
	}
	return wm.SendAndConfirmInstructions(feePayer.PublicKey(), instructions, fromWallets)
}

func (wm *WalletManager) makeTransferTransaction(from solana.PrivateKey, to solana.PublicKey, lamports uint64) (*solana.Transaction, error) {
	instruction := makeTransferInstruction(from.PublicKey(), to, lamports)
	recent, err := wm.Client.GetRecentBlockhash(context.TODO(), wm.Commitment)
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

func makeTransferInstruction(from, to solana.PublicKey, lamports uint64) solana.Instruction {
	return system.NewTransferInstructionBuilder().
		SetFundingAccount(from).
		SetRecipientAccount(to).
		SetLamports(lamports).
		Build()
}

func (wm *WalletManager) SendTokens(feePayer solana.PrivateKey, to, mint solana.PublicKey, amount uint64) (solana.Signature, error) {
	return wm.SendTokensTransaction(feePayer, []SendTokensInstructionParams{{feePayer, to, mint, amount}})
}

func (wm *WalletManager) SendTokensTransaction(feePayer solana.PrivateKey, instructionsParams []SendTokensInstructionParams) (solana.Signature, error) {
	var instructions []solana.Instruction
	var signers []solana.PrivateKey
	for _, params := range instructionsParams {
		processAddress := func(to solana.PublicKey) (solana.PublicKey, error) {
			atokAddress, fromAtokInst, err := wm.getOrCreateAssociatedTokenAddress(params.From, to, params.Mint)
			if err != nil {
				return solana.PublicKey{}, errors.Errorf(
					"failed to find associated token address for %s. err: %s",
					to.String(),
					err.Error(),
				)
			}
			if fromAtokInst != nil {
				instructions = append(instructions, fromAtokInst)
			}
			return atokAddress, nil
		}
		fromAssociatedAddress, err := processAddress(params.From.PublicKey())
		if err != nil {
			return solana.Signature{}, err
		}
		toAssociatedAddress, err := processAddress(params.To)
		if err != nil {
			return solana.Signature{}, err
		}
		instruction := token.NewTransferInstructionBuilder().
			SetAmount(params.Amount).
			SetSourceAccount(fromAssociatedAddress).
			SetDestinationAccount(toAssociatedAddress).
			SetOwnerAccount(params.From.PublicKey()).
			Build()
		instructions = append(instructions, instruction)
		signers = append(signers, params.From)
	}
	signers = appendSignerIfNotPresented(signers, feePayer)
	return wm.SendAndConfirmInstructions(
		feePayer.PublicKey(),
		instructions,
		signers,
	)
}

func (wm *WalletManager) getOrCreateAssociatedTokenAddress(
	payer solana.PrivateKey,
	account,
	mint solana.PublicKey,
) (solana.PublicKey, *atok.Instruction, error) {
	atokAddress, _, err := solana.FindAssociatedTokenAddress(account, mint)
	if err != nil {
		return solana.PublicKey{}, nil, err
	}
	_, err = wm.Client.GetAccountInfoWithOpts(context.TODO(), atokAddress, &rpc.GetAccountInfoOpts{
		Commitment: wm.Commitment,
	})
	var createInstruction *atok.Instruction
	if err != nil {
		createInstruction = atok.NewCreateInstructionBuilder().
			SetPayer(payer.PublicKey()).
			SetMint(mint).
			SetWallet(account).
			Build()
	}
	return atokAddress, createInstruction, nil
}

func (wm *WalletManager) SendAndConfirmInstructions(
	feePayer solana.PublicKey,
	instructions []solana.Instruction,
	signers []solana.PrivateKey,
) (solana.Signature, error) {
	recent, err := wm.Client.GetRecentBlockhash(wm.Context, wm.Commitment)
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
	return wm.SendAndConfirmTransaction(tx)
}

func (wm *WalletManager) SendAndConfirmTransaction(
	tx *solana.Transaction,
) (solana.Signature, error) {
	sig, err := wm.Client.SendTransactionWithOpts(wm.Context, tx, rpc.TransactionOpts{
		SkipPreflight:       wm.SkipPreflight,
		PreflightCommitment: wm.Commitment,
	})
	if err != nil {
		return solana.Signature{}, err
	}
	return wm.awaitSignaturesConfirmation([]solana.Signature{sig})
}

func (wm *WalletManager) awaitSignaturesConfirmation(
	signatures []solana.Signature,
) (solana.Signature, error) {
	if len(signatures) == 0 {
		return solana.Signature{}, errors.New("signatures array is empty")
	}
	after := time.After(wm.ConfirmationTimeout)
	ticker := time.NewTicker(wm.ConfirmationDelay)

	for {
		select {
		case <-ticker.C:
			result, err := wm.Client.GetSignatureStatuses(wm.Context, true, signatures...)
			if err == nil {
				for idx, res := range result.Value {
					if res.Err == nil && res.ConfirmationStatus == wm.ConfirmationStatusType {
						return signatures[idx], nil
					}
				}
			}
		case <-after:
			return solana.Signature{}, errors.New("timeout")
		}
	}
}
