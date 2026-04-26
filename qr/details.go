package qr

// Details holds the data needed to generate a QR code URL.
type Details struct {
	ChainID           int
	ContactAddress    string
	Decimals          int
	ReceipientAddress string
	Dollars           int
	Cents             int
}
