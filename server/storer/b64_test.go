package storer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/orange-cloudfoundry/terraform-secure-backend/server/storer"
)

var _ = Describe("B64", func() {
	var storer Storer
	BeforeEach(func() {
		storer = NewB64(storerRec)
		storerRec.Reset()
	})

	Context("Store", func() {
		It("should give encode data in b64", func() {
			err := storer.Store("foo", Str2ReadCloser("bar"))
			Expect(err).ToNot(HaveOccurred())

			Expect(storerRec.RetrieveString("foo")).To(Equal("YmFy"))
		})
	})

	Context("Retrieve", func() {
		It("should give back reader with decoded data", func() {
			err := storer.Store("foo", Str2ReadCloser("bar"))
			Expect(err).ToNot(HaveOccurred())

			r, err := storer.Retrieve("foo")
			Expect(err).ToNot(HaveOccurred())

			Expect(string(ReadCloserToBytes(r))).To(Equal("bar"))
		})
	})

	Context("Delete", func() {
		It("should let next storer delete it", func() {
			err := storer.Store("foo", Str2ReadCloser("bar"))
			Expect(err).ToNot(HaveOccurred())

			err = storer.Delete("foo")
			Expect(err).ToNot(HaveOccurred())

			Expect(storerRec.IsDeletedCall("foo")).To(BeTrue())
		})
	})
})
