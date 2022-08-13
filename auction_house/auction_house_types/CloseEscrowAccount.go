// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package auction_house_types

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// CloseEscrowAccount is the `closeEscrowAccount` instruction.
type CloseEscrowAccount struct {
	EscrowPaymentBump *uint8

	// [0] = [SIGNER] wallet
	//
	// [1] = [WRITE] escrowPaymentAccount
	//
	// [2] = [] auctionHouse
	//
	// [3] = [] systemProgram
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewCloseEscrowAccountInstructionBuilder creates a new `CloseEscrowAccount` instruction builder.
func NewCloseEscrowAccountInstructionBuilder() *CloseEscrowAccount {
	nd := &CloseEscrowAccount{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 4),
	}
	return nd
}

// SetEscrowPaymentBump sets the "escrowPaymentBump" parameter.
func (inst *CloseEscrowAccount) SetEscrowPaymentBump(escrowPaymentBump uint8) *CloseEscrowAccount {
	inst.EscrowPaymentBump = &escrowPaymentBump
	return inst
}

// SetWalletAccount sets the "wallet" account.
func (inst *CloseEscrowAccount) SetWalletAccount(wallet ag_solanago.PublicKey) *CloseEscrowAccount {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(wallet).SIGNER()
	return inst
}

// GetWalletAccount gets the "wallet" account.
func (inst *CloseEscrowAccount) GetWalletAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetEscrowPaymentAccountAccount sets the "escrowPaymentAccount" account.
func (inst *CloseEscrowAccount) SetEscrowPaymentAccountAccount(escrowPaymentAccount ag_solanago.PublicKey) *CloseEscrowAccount {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(escrowPaymentAccount).WRITE()
	return inst
}

// GetEscrowPaymentAccountAccount gets the "escrowPaymentAccount" account.
func (inst *CloseEscrowAccount) GetEscrowPaymentAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

// SetAuctionHouseAccount sets the "auctionHouse" account.
func (inst *CloseEscrowAccount) SetAuctionHouseAccount(auctionHouse ag_solanago.PublicKey) *CloseEscrowAccount {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(auctionHouse)
	return inst
}

// GetAuctionHouseAccount gets the "auctionHouse" account.
func (inst *CloseEscrowAccount) GetAuctionHouseAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

// SetSystemProgramAccount sets the "systemProgram" account.
func (inst *CloseEscrowAccount) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *CloseEscrowAccount {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
func (inst *CloseEscrowAccount) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(3)
}

func (inst CloseEscrowAccount) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_CloseEscrowAccount,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst CloseEscrowAccount) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CloseEscrowAccount) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.EscrowPaymentBump == nil {
			return errors.New("EscrowPaymentBump parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Wallet is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.EscrowPaymentAccount is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.AuctionHouse is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
	}
	return nil
}

func (inst *CloseEscrowAccount) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("CloseEscrowAccount")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=1]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("EscrowPaymentBump", *inst.EscrowPaymentBump))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=4]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("       wallet", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("escrowPayment", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta(" auctionHouse", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(ag_format.Meta("systemProgram", inst.AccountMetaSlice.Get(3)))
					})
				})
		})
}

func (obj CloseEscrowAccount) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `EscrowPaymentBump` param:
	err = encoder.Encode(obj.EscrowPaymentBump)
	if err != nil {
		return err
	}
	return nil
}
func (obj *CloseEscrowAccount) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `EscrowPaymentBump`:
	err = decoder.Decode(&obj.EscrowPaymentBump)
	if err != nil {
		return err
	}
	return nil
}

// NewCloseEscrowAccountInstruction declares a new CloseEscrowAccount instruction with the provided parameters and accounts.
func NewCloseEscrowAccountInstruction(
	// Parameters:
	escrowPaymentBump uint8,
	// Accounts:
	wallet ag_solanago.PublicKey,
	escrowPaymentAccount ag_solanago.PublicKey,
	auctionHouse ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey) *CloseEscrowAccount {
	return NewCloseEscrowAccountInstructionBuilder().
		SetEscrowPaymentBump(escrowPaymentBump).
		SetWalletAccount(wallet).
		SetEscrowPaymentAccountAccount(escrowPaymentAccount).
		SetAuctionHouseAccount(auctionHouse).
		SetSystemProgramAccount(systemProgram)
}
