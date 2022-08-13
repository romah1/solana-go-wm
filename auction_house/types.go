package auction_house

import (
	"github.com/gagliardetto/solana-go"
	"solana-go-wm/auction_house/auction_house_types"
	solana_go_wm "solana-go-wm/wallet_manager"
)

type AuctionHouseActor struct {
	Wm                  *solana_go_wm.WalletManager
	AuctionHouseAccount solana.PublicKey
	AuctionHouseData    auction_house_types.AuctionHouse
}

type AuctionHouseBuyData struct {
	Owner       solana.PublicKey
	MintAddress solana.PublicKey
	MintAta     solana.PublicKey
	Price       int
	TokenSize   int
	Creators    []solana.PublicKey
}
