package storer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/orange-cloudfoundry/terraform-secure-backend/server/credhub/credhubfakes"
	. "github.com/orange-cloudfoundry/terraform-secure-backend/server/storer"
)

var _ = Describe("Credhub", func() {
	var storer Storer
	var fakeClient *credhubfakes.FakeCredhubClient
	BeforeEach(func() {
		fakeClient = new(credhubfakes.FakeCredhubClient)
		storer = NewCredhub(fakeClient)
	})

	Context("Store", func() {
		It("should give store data to credhub in json", func() {
			err := storer.Store("foo", Str2ReadCloser(`{"foo": "bar"}`))
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeClient.SetJSONCallCount()).Should(Equal(1))
		})
	})

	Context("Retrieve", func() {
		It("should give back reader with decoded data", func() {
			err := storer.Store("foo", Str2ReadCloser(`{"foo": "bar"}`))
			Expect(err).ToNot(HaveOccurred())

			_, err = storer.Retrieve("foo")
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeClient.GetLatestJSONCallCount()).Should(Equal(1))
		})
	})

	Context("Delete", func() {
		It("should call delete on credhub", func() {
			err := storer.Store("foo", Str2ReadCloser(`{"foo": "bar"}`))
			Expect(err).ToNot(HaveOccurred())

			err = storer.Delete("foo")
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeClient.DeleteCallCount()).Should(Equal(1))
		})
	})
})
