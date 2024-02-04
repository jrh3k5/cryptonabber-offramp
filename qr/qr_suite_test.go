package qr_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestQr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Qr Suite")
}
