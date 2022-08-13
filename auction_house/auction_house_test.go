package auction_house

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"solana-go-wm/wallet_manager"
	"testing"
)

var wm = wallet_manager.NewWalletManager(rpc.New(rpc.MainNetBeta_RPC))
var aucHouse, _ = NewAuctionHouseActor(wm, CoralCubeAuctionHouseAccount)

func TestAuctionHouseActor_Sell(t *testing.T) {
	privateKeyString := ""
	mintString := ""

	if privateKeyString == "" || mintString == "" {
		t.Skip("privateKeyString or mintString is not set")
	}

	sig, err := aucHouse.Sell(
		solana.MustPrivateKeyFromBase58(privateKeyString),
		solana.MustPublicKeyFromBase58(mintString),
		uint64(0.01*float64(solana.LAMPORTS_PER_SOL)),
		1,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(sig.String())
}
