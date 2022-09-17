package wallet_manager

import "github.com/gagliardetto/solana-go"

func appendSignerIfNotPresented(signers []solana.PrivateKey, newSigner solana.PrivateKey) []solana.PrivateKey {
	for _, signer := range signers {
		if signer.PublicKey() == newSigner.PublicKey() {
			return signers
		}
	}
	return append(signers, newSigner)
}
