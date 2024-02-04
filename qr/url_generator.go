package qr

import "context"

// URLGenerator is used to generate a URL for a QR code.
type URLGenerator interface {
	// Generate generates a URL to be presented for a QR code
	Generate(ctx context.Context, qrDetails *Details) (string, error)
}
