package auction_house

import (
	"context"
	"encoding/binary"
	bin "github.com/gagliardetto/binary"
	token_metadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
	"solana-go-wm/auction_house/auction_house_types"
	"solana-go-wm/wallet_manager"
)

func NewAuctionHouseActor(wm *wallet_manager.WalletManager, auctionHouseAccount solana.PublicKey) (*AuctionHouseActor, error) {
	aucHouseData, err := getAuctionHouseAccountData(wm.Client, auctionHouseAccount)
	if err != nil {
		return nil, errors.Errorf("failed to get auc house account data. err: %s", err.Error())
	}
	return &AuctionHouseActor{
		Wm:                  wm,
		AuctionHouseAccount: auctionHouseAccount,
		AuctionHouseData:    aucHouseData,
	}, nil
}

func (aucHouse *AuctionHouseActor) Buy(buyer solana.PrivateKey, data AuctionHouseBuyData) (solana.Signature, error) {
	buyerEscrowAccount, buyerEscrowBump, err := aucHouse.getBuyerEscrow(buyer.PublicKey())
	if err != nil {
		return solana.Signature{}, err
	}
	buyerTradeStateAccount, buyerTradeStateBump, err := aucHouse.getTradeState(
		buyer.PublicKey(),
		data.MintAta,
		data.MintAddress,
		data.Price,
		data.TokenSize,
	)
	if err != nil {
		return solana.Signature{}, err
	}
	tokenWallet, err := getTokenWallet(buyer.PublicKey(), data.MintAddress)
	if err != nil {
		return solana.Signature{}, err
	}
	metadata, err := getMetadata(data.MintAddress)
	if err != nil {
		return solana.Signature{}, err
	}
	buyInstruction := auction_house_types.NewBuyInstructionBuilder().
		SetTradeStateBump(buyerTradeStateBump).
		SetEscrowPaymentBump(buyerEscrowBump).
		SetBuyerPrice(data.Price).
		SetTokenSize(data.TokenSize).
		SetWalletAccount(buyer.PublicKey()).
		SetPaymentAccountAccount(buyer.PublicKey()).
		SetTransferAuthorityAccount(solana.SystemProgramID).
		SetMetadataAccount(metadata).
		SetTokenAccountAccount(data.MintAta).
		SetEscrowPaymentAccountAccount(buyerEscrowAccount).
		SetTreasuryMintAccount(aucHouse.AuctionHouseData.TreasuryMint).
		SetAuthorityAccount(aucHouse.AuctionHouseData.Authority).
		SetAuctionHouseAccount(aucHouse.AuctionHouseAccount).
		SetAuctionHouseFeeAccountAccount(aucHouse.AuctionHouseData.AuctionHouseFeeAccount).
		SetBuyerTradeStateAccount(buyerTradeStateAccount).
		SetTokenProgramAccount(solana.TokenProgramID).
		SetSystemProgramAccount(solana.SystemProgramID).
		SetRentAccount(solana.SysVarRentPubkey).
		Build()
	sellerTradeStateAccount, _, err := aucHouse.getTradeState(
		data.Owner,
		data.MintAta,
		data.MintAddress,
		data.Price,
		data.TokenSize,
	)
	if err != nil {
		return solana.Signature{}, err
	}
	freeTradeStateAccount, freeTradeStateAccountBump, err := aucHouse.getTradeState(
		data.Owner,
		data.MintAta,
		data.MintAddress,
		0,
		data.TokenSize,
	)
	if err != nil {
		return solana.Signature{}, err
	}
	programAsSignerAccount, programAsSignerBump, err := getProgramAsSigner()
	if err != nil {
		return solana.Signature{}, err
	}
	executeSaleInstructionBuilder := auction_house_types.NewExecuteSaleInstructionBuilder().
		SetEscrowPaymentBump(buyerEscrowBump).
		SetFreeTradeStateBump(freeTradeStateAccountBump).
		SetProgramAsSignerBump(programAsSignerBump).
		SetBuyerPrice(data.Price).
		SetTokenSize(data.TokenSize).
		SetBuyerAccount(buyer.PublicKey()).
		SetSellerAccount(data.Owner).
		SetMetadataAccount(metadata).
		SetTokenAccountAccount(data.MintAta).
		SetTokenMintAccount(data.MintAddress).
		SetEscrowPaymentAccountAccount(buyerEscrowAccount).
		SetTreasuryMintAccount(aucHouse.AuctionHouseData.TreasuryMint).
		SetSellerPaymentReceiptAccountAccount(data.Owner).
		SetBuyerReceiptTokenAccountAccount(tokenWallet).
		SetAuthorityAccount(aucHouse.AuctionHouseData.Authority).
		SetAuctionHouseAccount(aucHouse.AuctionHouseAccount).
		SetAuctionHouseFeeAccountAccount(aucHouse.AuctionHouseData.AuctionHouseFeeAccount).
		SetAuctionHouseTreasuryAccount(aucHouse.AuctionHouseData.AuctionHouseTreasury).
		SetSellerTradeStateAccount(sellerTradeStateAccount).
		SetBuyerTradeStateAccount(buyerTradeStateAccount).
		SetTokenProgramAccount(solana.TokenProgramID).
		SetSystemProgramAccount(solana.SystemProgramID).
		SetAtaProgramAccount(solana.SPLAssociatedTokenAccountProgramID).
		SetProgramAsSignerAccount(programAsSignerAccount).
		SetRentAccount(solana.SysVarRentPubkey).
		SetFreeTradeStateAccount(freeTradeStateAccount)

	for _, creator := range data.Creators {
		executeSaleInstructionBuilder.Append(solana.NewAccountMeta(creator, true, false))
	}
	return aucHouse.Wm.SendAndConfirmInstructions(
		buyer.PublicKey(),
		[]solana.Instruction{buyInstruction, executeSaleInstructionBuilder.Build()},
		[]solana.PrivateKey{buyer},
	)
}

func (aucHouse *AuctionHouseActor) Sell(
	seller solana.PrivateKey,
	mint solana.PublicKey,
	priceLamports uint64,
	amount uint64,
) (solana.Signature, error) {
	meta, err := getMetadata(mint)
	if err != nil {
		return solana.Signature{}, err
	}
	mintAta, _, err := solana.FindAssociatedTokenAddress(seller.PublicKey(), mint)
	if err != nil {
		return solana.Signature{}, err
	}
	programAsSigner, programAsSignerBump, err := getProgramAsSigner()
	if err != nil {
		return solana.Signature{}, err
	}
	tradeState, tradeBump, err := aucHouse.getTradeState(seller.PublicKey(), mintAta, mint, priceLamports, amount)
	if err != nil {
		return solana.Signature{}, err
	}
	freeTradeState, freeTradeBump, err := aucHouse.getTradeState(seller.PublicKey(), mintAta, mint, 0, amount)
	if err != nil {
		return solana.Signature{}, err
	}
	instruction := auction_house_types.NewSellInstructionBuilder().
		SetTradeStateBump(tradeBump).
		SetFreeTradeStateBump(freeTradeBump).
		SetProgramAsSignerBump(programAsSignerBump).
		SetBuyerPrice(priceLamports).
		SetTokenSize(amount).
		SetWalletAccount(seller.PublicKey()).
		SetMetadataAccount(meta).
		SetTokenAccountAccount(mintAta).
		SetAuthorityAccount(aucHouse.AuctionHouseData.Authority).
		SetAuctionHouseAccount(aucHouse.AuctionHouseAccount).
		SetAuctionHouseFeeAccountAccount(aucHouse.AuctionHouseData.AuctionHouseFeeAccount).
		SetSellerTradeStateAccount(tradeState).
		SetFreeSellerTradeStateAccount(freeTradeState).
		SetTokenProgramAccount(solana.TokenProgramID).
		SetSystemProgramAccount(solana.SystemProgramID).
		SetProgramAsSignerAccount(programAsSigner).
		SetRentAccount(solana.SysVarRentPubkey).
		Build()
	return aucHouse.Wm.SendAndConfirmInstructions(seller.PublicKey(), []solana.Instruction{instruction}, []solana.PrivateKey{seller})
}

func (aucHouse *AuctionHouseActor) getBuyerEscrow(wallet solana.PublicKey) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{[]byte(auctionHouse), aucHouse.AuctionHouseAccount.Bytes(), wallet.Bytes()},
		auction_house_types.ProgramID,
	)
}

func (aucHouse *AuctionHouseActor) getTradeState(
	wallet,
	tokenAccount,
	tokenMint solana.PublicKey,
	price uint64,
	tokenSize uint64,
) (solana.PublicKey, uint8, error) {
	buyPriceBytes := make([]byte, 8)
	tokenSizeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(tokenSizeBytes, tokenSize)
	binary.LittleEndian.PutUint64(buyPriceBytes, price)
	return solana.FindProgramAddress(
		[][]byte{
			[]byte(auctionHouse),
			wallet.Bytes(),
			aucHouse.AuctionHouseAccount.Bytes(),
			tokenAccount.Bytes(),
			aucHouse.AuctionHouseData.TreasuryMint.Bytes(),
			tokenMint.Bytes(),
			buyPriceBytes,
			tokenSizeBytes,
		},
		auction_house_types.ProgramID,
	)
}

func getTokenWallet(wallet solana.PublicKey, mint solana.PublicKey) (solana.PublicKey, error) {
	addr, _, err := solana.FindProgramAddress(
		[][]byte{
			wallet.Bytes(),
			solana.TokenProgramID.Bytes(),
			mint.Bytes(),
		},
		solana.SPLAssociatedTokenAccountProgramID,
	)
	return addr, err
}

func getMetadata(mint solana.PublicKey) (solana.PublicKey, error) {
	addr, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("metadata"),
			token_metadata.ProgramID.Bytes(),
			mint.Bytes(),
		},
		token_metadata.ProgramID,
	)
	return addr, err
}

func getAuctionHouseAccountData(client *rpc.Client, auctionHouseAccountKey solana.PublicKey) (auction_house_types.AuctionHouse, error) {
	candyMachineRaw, err := client.GetAccountInfo(context.TODO(), auctionHouseAccountKey)
	if err != nil {
		return auction_house_types.AuctionHouse{}, err
	}
	dec := bin.NewBorshDecoder(candyMachineRaw.Value.Data.GetBinary())
	var aucHouse auction_house_types.AuctionHouse
	err = dec.Decode(&aucHouse)
	if err != nil {
		return auction_house_types.AuctionHouse{}, err
	}
	return aucHouse, nil
}

func getProgramAsSigner() (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{[]byte(auctionHouse), []byte("signer")},
		auction_house_types.ProgramID,
	)
}
