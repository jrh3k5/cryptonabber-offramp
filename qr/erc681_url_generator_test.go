package qr_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jrh3k5/cryptonabber-offramp/v3/qr"
)

var _ = Describe("Erc681UrlGenerator", func() {
	var generator *qr.ERC681URLGenerator
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
		generator = qr.NewERC681URLGenerator()
	})

	Context("Generate", func() {
		It("generates a valid ERC681 URL", func() {
			details := &qr.Details{
				ChainID:           8453,
				ContactAddress:    "0x833589fcd6edb6e08f4c7c32d4f71b54bda02913",
				Decimals:          6,
				ReceipientAddress: "0x407DF19995bBA21E71EC6e6b72FEba70318031Be",
				Dollars:           1,
				Cents:             28,
			}

			url, err := generator.Generate(ctx, details)
			Expect(err).ToNot(HaveOccurred(), "generating the URL should not fail")
			Expect(url).To(Equal(url), "the correct URL should be generated")
		})
	})
})
