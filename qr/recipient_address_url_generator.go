package qr

import "context"

// RecipientAddressURLGenerator is a generator that just generates
// a URL to show the given recipient address. This is useful for wallets
// that only expect wallet addresses in the QR code, rather than full-formed
// URLs such as ERC-681 URLs.
type RecipientAddressURLGenerator struct {
}

func NewRecipientAddressURLGenerator() *RecipientAddressURLGenerator {
	return &RecipientAddressURLGenerator{}
}

func (*RecipientAddressURLGenerator) Generate(ctx context.Context, qrDetails *Details) (string, error) {
	return qrDetails.ReceipientAddress, nil
}
