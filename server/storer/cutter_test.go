package storer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/orange-cloudfoundry/terraform-secure-backend/server/storer"
)

var _ = Describe("Cutter", func() {
	var storer Storer
	BeforeEach(func() {
		storer = NewCutter(storerRec, 1)
		storerRec.Reset()
	})

	Context("Store", func() {
		It("should store part as json without exceed chunk size", func() {
			err := storer.Store("foo", Str2ReadCloser("012"))
			Expect(err).ToNot(HaveOccurred())

			Expect(storerRec.RetrieveIndex("foo").NumParts).To(Equal(3))
			Expect(storerRec.RetrievePart("foo/0")).To(Equal("0"))
			Expect(storerRec.RetrievePart("foo/1")).To(Equal("1"))
			Expect(storerRec.RetrievePart("foo/2")).To(Equal("2"))
		})
	})

	Context("Retrieve", func() {
		It("should give back reader with decoded data", func() {
			err := storer.Store("foo", Str2ReadCloser("012"))
			Expect(err).ToNot(HaveOccurred())

			r, err := storer.Retrieve("foo")
			Expect(err).ToNot(HaveOccurred())

			Expect(string(ReadCloserToBytes(r))).To(Equal("012"))
		})
	})

	Context("Delete", func() {
		It("should delete all part by passing path to next storer", func() {
			err := storer.Store("foo", Str2ReadCloser("012"))
			Expect(err).ToNot(HaveOccurred())

			err = storer.Delete("foo")
			Expect(err).ToNot(HaveOccurred())

			Expect(storerRec.IsDeletedCall("foo/0")).To(BeTrue())
			Expect(storerRec.IsDeletedCall("foo/1")).To(BeTrue())
			Expect(storerRec.IsDeletedCall("foo/2")).To(BeTrue())
			Expect(storerRec.IsDeletedCall("foo/index")).To(BeTrue())
		})
	})
})
