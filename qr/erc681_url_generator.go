package qr

import (
	"context"
	"fmt"
	"math"
)

// ERC681URLGenerator is a generator that generates URLs in compliance with
// ERC20 transfer for ERC-681: https://eips.ethereum.org/EIPS/eip-681
type ERC681URLGenerator struct {
}

// NewERC681URLGenerator
func NewERC681URLGenerator() *ERC681URLGenerator {
	return &ERC681URLGenerator{}
}

func (*ERC681URLGenerator) Generate(ctx context.Context, qrDetails *Details) (string, error) {
	decimals := qrDetails.Decimals
	tokenDollars := qrDetails.Dollars * int(math.Pow10(decimals))
	tokenCents := qrDetails.Cents * int(math.Pow10(decimals-2))
	tokenAmount := tokenCents + tokenDollars

	url := fmt.Sprintf("ethereum:%s@%d/transfer?address=%s&uint256=%d", qrDetails.ContactAddress, qrDetails.ChainID, qrDetails.ReceipientAddress, tokenAmount)

	return url, nil
}
