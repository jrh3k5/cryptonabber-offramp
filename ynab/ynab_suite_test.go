package ynab_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestYnab(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ynab Suite")
}
