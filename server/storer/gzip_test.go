package storer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/orange-cloudfoundry/terraform-secure-backend/server/storer"
)

var _ = Describe("Gzip", func() {
	var storer Storer
	BeforeEach(func() {
		storer = NewGzip(storerRec)
		storerRec.Reset()
	})

	Context("Store", func() {
		It("should give encode data in gzip", func() {
			err := storer.Store("foo", Str2ReadCloser("A"))
			Expect(err).ToNot(HaveOccurred())

			Expect(storerRec.RetrieveBytes("foo")).To(Equal([]byte{31, 139, 8, 0, 0, 0, 0, 0, 2, 255, 114, 4, 4, 0, 0, 255, 255, 139, 158, 217, 211, 1, 0, 0, 0}))
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
