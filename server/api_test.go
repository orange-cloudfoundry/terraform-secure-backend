package server_test

import (
	. "github.com/orange-cloudfoundry/terraform-secure-backend/server"

	"bytes"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials/values"
	"encoding/json"
	"errors"
	"github.com/hashicorp/terraform/state"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/orange-cloudfoundry/terraform-secure-backend/server/serverfakes"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Api", func() {
	log.SetOutput(ioutil.Discard)
	var fakeClient *serverfakes.FakeCredhubClient
	var apiController *ApiController
	var lockStore *LockStore
	var responseRecorder *httptest.ResponseRecorder
	BeforeEach(func() {
		responseRecorder = httptest.NewRecorder()
		fakeClient = new(serverfakes.FakeCredhubClient)
		lockStore = NewLockStore(fakeClient)
		apiController = NewApiController("test", fakeClient, lockStore)
	})
	Context("Store", func() {
		It("should store data when giving state", func() {
			apiController.Store(responseRecorder, httptest.NewRequest("POST", "http://fakeurl.com", bytes.NewBufferString(`{"key": "value"}`)))
			Expect(fakeClient.SetJSONCallCount()).Should(Equal(1))
			Expect(responseRecorder.Code).Should(Equal(http.StatusOK))
		})
		It("should panic if unmarshal was in error", func() {
			Expect(func() {
				apiController.Store(
					responseRecorder,
					httptest.NewRequest("POST", "http://fakeurl.com", bytes.NewBufferString("")))
			}).Should(Panic())
		})
		It("should panic if setting credential was in error", func() {
			fakeClient.SetJSONReturns(credentials.JSON{}, errors.New("a fake error"))
			Expect(func() {
				apiController.Store(responseRecorder, httptest.NewRequest("POST", "http://fakeurl.com", bytes.NewBufferString(`{"key": "value"}`)))
			}).Should(Panic())
		})
	})
	Context("Retrieve", func() {
		It("should giving data from credhub when exists", func() {
			data := values.JSON{
				"key": "value",
			}
			fakeClient.GetLatestJSONReturns(credentials.JSON{
				Value: data,
			}, nil)

			apiController.Retrieve(responseRecorder, httptest.NewRequest("GET", "http://fakeurl.com", nil))

			Expect(fakeClient.GetLatestJSONCallCount()).Should(Equal(1))
			Expect(responseRecorder.Code).Should(Equal(http.StatusOK))

			var resultData map[string]interface{}
			err := json.Unmarshal(responseRecorder.Body.Bytes(), &resultData)
			Expect(err).ToNot(HaveOccurred())

			Expect(resultData).Should(HaveKey("key"))
			Expect(resultData["key"]).Should(Equal("value"))
		})
		It("should answer with http code status no content when there is no data in credhub", func() {

			fakeClient.GetLatestJSONReturns(credentials.JSON{}, errors.New("does not exist"))

			apiController.Retrieve(responseRecorder, httptest.NewRequest("GET", "http://fakeurl.com", nil))

			Expect(fakeClient.GetLatestJSONCallCount()).Should(Equal(1))
			Expect(responseRecorder.Code).Should(Equal(http.StatusNoContent))
		})
		It("should panic if getting credential was in error", func() {
			fakeClient.GetLatestJSONReturns(credentials.JSON{}, errors.New("a fake error"))
			Expect(func() {
				apiController.Retrieve(responseRecorder, httptest.NewRequest("GET", "http://fakeurl.com", nil))
			}).Should(Panic())
		})
	})
	Context("Delete", func() {
		It("should delete data from credhub and delete lock", func() {
			req := httptest.NewRequest("DELETE", "http://fakeurl.com", nil)
			apiController.Delete(responseRecorder, req)

			Expect(fakeClient.DeleteCallCount()).Should(Equal(2))
			Expect(fakeClient.DeleteArgsForCall(0)).Should(Equal(apiController.CredhubName(req)))
			Expect(fakeClient.DeleteArgsForCall(1)).Should(Equal(apiController.CredhubName(req) + LOCK_SUFFIX))
			Expect(responseRecorder.Code).Should(Equal(http.StatusOK))
		})
		It("should panic if deleting lock was in error", func() {
			req := httptest.NewRequest("DELETE", "http://fakeurl.com", nil)
			fakeClient.DeleteReturnsOnCall(1, errors.New("fake error"))
			Expect(func() {
				apiController.Delete(responseRecorder, req)
			}).Should(Panic())
		})
		It("should panic if deleting data was in error", func() {
			req := httptest.NewRequest("DELETE", "http://fakeurl.com", nil)
			fakeClient.DeleteReturnsOnCall(0, errors.New("fake error"))
			Expect(func() {
				apiController.Delete(responseRecorder, req)
			}).Should(Panic())
		})
	})
	Context("Lock", func() {
		It("should lock when state info is passed and not already locked", func() {
			fakeClient.GetLatestValueReturns(credentials.Value{}, errors.New("does not exist"))
			req := httptest.NewRequest("LOCK", "http://fakeurl.com", bytes.NewBuffer((&state.LockInfo{
				ID: "fakeid",
			}).Marshal()))

			apiController.Lock(responseRecorder, req)

			Expect(fakeClient.SetValueCallCount()).Should(Equal(1))
			Expect(responseRecorder.Code).Should(Equal(http.StatusOK))
		})
		It("should return http code locked and lock info if it's already locked", func() {
			req := httptest.NewRequest("LOCK", "http://fakeurl.com", bytes.NewBuffer((&state.LockInfo{
				ID: "fakeid",
			}).Marshal()))
			fakeClient.GetLatestValueReturns(credentials.Value{
				Value: values.Value("an id"),
			}, nil)

			apiController.Lock(responseRecorder, req)

			Expect(responseRecorder.Code).Should(Equal(http.StatusLocked))
			var lockInfo state.LockInfo
			err := json.Unmarshal(responseRecorder.Body.Bytes(), &lockInfo)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(lockInfo.ID).Should(Equal("an id"))
		})
		It("should panic if unmarshal was in error", func() {
			req := httptest.NewRequest("LOCK", "http://fakeurl.com", bytes.NewBufferString(""))
			fakeClient.GetLatestValueReturns(credentials.Value{}, errors.New("does not exist"))

			Expect(func() {
				apiController.Lock(responseRecorder, req)
			}).Should(Panic())
		})
		It("should panic if locking was in error", func() {
			req := httptest.NewRequest("LOCK", "http://fakeurl.com", bytes.NewBuffer((&state.LockInfo{
				ID: "fakeid",
			}).Marshal()))
			fakeClient.SetValueReturns(credentials.Value{}, errors.New("fake error"))
			fakeClient.GetLatestValueReturns(credentials.Value{}, errors.New("does not exist"))

			Expect(func() {
				apiController.Lock(responseRecorder, req)
			}).Should(Panic())
		})
	})
	Context("UnLock", func() {
		It("should unlock when state info is passed with correct id", func() {
			id := "myid"
			fakeClient.GetLatestValueReturns(credentials.Value{
				Value: values.Value(id),
			}, nil)
			req := httptest.NewRequest("LOCK", "http://fakeurl.com", bytes.NewBuffer((&state.LockInfo{
				ID: id,
			}).Marshal()))

			apiController.UnLock(responseRecorder, req)

			Expect(fakeClient.DeleteCallCount()).Should(Equal(1))
			Expect(responseRecorder.Code).Should(Equal(http.StatusOK))
		})
		It("should return http code conflict and lock info if it's already locked and lock info id is not the one expected", func() {
			id := "myid"
			fakeClient.GetLatestValueReturns(credentials.Value{
				Value: values.Value(id),
			}, nil)
			req := httptest.NewRequest("LOCK", "http://fakeurl.com", bytes.NewBuffer((&state.LockInfo{
				ID: "otherid",
			}).Marshal()))

			apiController.UnLock(responseRecorder, req)

			Expect(responseRecorder.Code).Should(Equal(http.StatusConflict))
			var lockInfo state.LockInfo
			err := json.Unmarshal(responseRecorder.Body.Bytes(), &lockInfo)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(lockInfo.ID).Should(Equal(id))
		})
		It("should panic if unmarshal was in error", func() {
			req := httptest.NewRequest("UNLOCK", "http://fakeurl.com", bytes.NewBufferString(""))

			Expect(func() {
				apiController.UnLock(responseRecorder, req)
			}).Should(Panic())
		})
		It("should panic if unlocking was in error", func() {
			id := "myid"
			fakeClient.GetLatestValueReturns(credentials.Value{
				Value: values.Value(id),
			}, nil)
			req := httptest.NewRequest("LOCK", "http://fakeurl.com", bytes.NewBuffer((&state.LockInfo{
				ID: id,
			}).Marshal()))
			fakeClient.DeleteReturns(errors.New("fake error"))
			Expect(func() {
				apiController.UnLock(responseRecorder, req)
			}).Should(Panic())
		})
	})
	Context("List", func() {
		It("should give a list credentials", func() {
			req := httptest.NewRequest("GET", "http://fakeurl.com", bytes.NewBufferString(""))
			fakeClient.FindByPathReturns(credentials.FindResults{[]credentials.Base{
				{
					Name:             apiController.CredhubName(req) + "data1",
					VersionCreatedAt: "now",
				},
				{
					Name:             apiController.CredhubName(req) + "data2",
					VersionCreatedAt: "now",
				},
				{
					Name:             apiController.CredhubName(req) + LOCK_SUFFIX,
					VersionCreatedAt: "now",
				},
			}}, nil)
			fakeClient.GetLatestValueReturnsOnCall(0, credentials.Value{}, errors.New("does not exist"))
			fakeClient.GetLatestValueReturnsOnCall(1, credentials.Value{
				Value: values.Value("id"),
			}, nil)

			apiController.List(responseRecorder, req)

			Expect(responseRecorder.Code).Should(Equal(http.StatusOK))
			var creds []CredModel
			err := json.Unmarshal(responseRecorder.Body.Bytes(), &creds)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(creds).Should(HaveLen(2))

			Expect(creds[0].Name).Should(Equal("data1"))
			Expect(creds[0].IsLocked).Should(BeFalse())

			Expect(creds[1].Name).Should(Equal("data2"))
			Expect(creds[1].IsLocked).Should(BeTrue())
			Expect(creds[1].CurrentLockId).Should(Equal("id"))
		})
		It("should panic if find was in error", func() {
			fakeClient.FindByPathReturns(credentials.FindResults{[]credentials.Base{}}, errors.New("a fake error"))
			req := httptest.NewRequest("GET", "http://fakeurl.com", nil)
			Expect(func() {
				apiController.List(responseRecorder, req)
			}).Should(Panic())
		})
	})
})
