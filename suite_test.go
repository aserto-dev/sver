package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	ginkgoT *testing.T
)

func TestCalcVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	ginkgoT = t
	RunSpecs(t, "sver suite")
}
