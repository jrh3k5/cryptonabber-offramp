package qr_test

import (
	"context"

	"github.com/jrh3k5/cryptonabber-offramp/v3/qr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RecipientAddressURLGenerator", func() {
	var generator *qr.RecipientAddressURLGenerator
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
		generator = qr.NewRecipientAddressURLGenerator()
	})

	Context("Generate", func() {
		It("just returns the given recipient address", func() {
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
			Expect(url).To(Equal(details.ReceipientAddress), "the URL should merely be the receipient address")
		})
	})
})
