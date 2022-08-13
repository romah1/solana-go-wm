// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package auction_house_types

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// CancelListingReceipt is the `cancelListingReceipt` instruction.
type CancelListingReceipt struct {

	// [0] = [WRITE] receipt
	//
	// [1] = [] systemProgram
	//
	// [2] = [] instruction
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewCancelListingReceiptInstructionBuilder creates a new `CancelListingReceipt` instruction builder.
func NewCancelListingReceiptInstructionBuilder() *CancelListingReceipt {
	nd := &CancelListingReceipt{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 3),
	}
	return nd
}

// SetReceiptAccount sets the "receipt" account.
func (inst *CancelListingReceipt) SetReceiptAccount(receipt ag_solanago.PublicKey) *CancelListingReceipt {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(receipt).WRITE()
	return inst
}

// GetReceiptAccount gets the "receipt" account.
func (inst *CancelListingReceipt) GetReceiptAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetSystemProgramAccount sets the "systemProgram" account.
func (inst *CancelListingReceipt) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *CancelListingReceipt {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
func (inst *CancelListingReceipt) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

// SetInstructionAccount sets the "instruction" account.
func (inst *CancelListingReceipt) SetInstructionAccount(instruction ag_solanago.PublicKey) *CancelListingReceipt {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(instruction)
	return inst
}

// GetInstructionAccount gets the "instruction" account.
func (inst *CancelListingReceipt) GetInstructionAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

func (inst CancelListingReceipt) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_CancelListingReceipt,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst CancelListingReceipt) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CancelListingReceipt) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Receipt is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.Instruction is not set")
		}
	}
	return nil
}

func (inst *CancelListingReceipt) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("CancelListingReceipt")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=0]").ParentFunc(func(paramsBranch ag_treeout.Branches) {})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=3]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("      receipt", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("systemProgram", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta("  instruction", inst.AccountMetaSlice.Get(2)))
					})
				})
		})
}

func (obj CancelListingReceipt) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}
func (obj *CancelListingReceipt) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewCancelListingReceiptInstruction declares a new CancelListingReceipt instruction with the provided parameters and accounts.
func NewCancelListingReceiptInstruction(
	// Accounts:
	receipt ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey,
	instruction ag_solanago.PublicKey) *CancelListingReceipt {
	return NewCancelListingReceiptInstructionBuilder().
		SetReceiptAccount(receipt).
		SetSystemProgramAccount(systemProgram).
		SetInstructionAccount(instruction)
}
