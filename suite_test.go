package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCalcVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "sver suite")
}
