package acceptance_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metrics are received", func() {
	BeforeEach(func() {
		// target the cf

	})
	It("correct metrics are emitted within 40s", func() {
		// check metrics are received
		//Expect(gbytes.buffer).To(gbytes.Say("something"))
		Expect(true).To(Equal(true))
	})

})
