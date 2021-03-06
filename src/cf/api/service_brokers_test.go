package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	"testing"
)

func TestServiceBrokersFindAll(t *testing.T) {
	responseBody := `{
  "resources": [
  	{
  	  "metadata": {
  	    "guid":"found-guid-1"
  	  },
  	  "entity": {
  	    "name": "found-name-1",
  	    "broker_url": "http://found.example.com-1",
  	    "auth_username": "found-username-1",
  	    "auth_password": "found-password-1"
  	  }
  	},
  	{
  	  "metadata": {
  	    "guid":"found-guid-2"
  	  },
  	  "entity": {
  	    "name": "found-name-2",
  	    "broker_url": "http://found.example.com-2",
  	    "auth_username": "found-username-2",
  	    "auth_password": "found-password-2"
  	  }
  	}
  ]
}`

	endpoint, status := testapi.CreateCheckableEndpoint(
		"GET",
		"/v2/service_brokers",
		nil,
		testapi.TestResponse{Status: http.StatusOK, Body: responseBody},
	)

	repo, ts := createServiceBrokerRepo(endpoint)
	defer ts.Close()

	serviceBrokers, apiResponse := repo.FindAll()

	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, len(serviceBrokers), 2)

	assert.Equal(t, serviceBrokers[0].Name, "found-name-1")
	assert.Equal(t, serviceBrokers[0].Guid, "found-guid-1")
	assert.Equal(t, serviceBrokers[0].Url, "http://found.example.com-1")
	assert.Equal(t, serviceBrokers[0].Username, "found-username-1")
	assert.Equal(t, serviceBrokers[0].Password, "found-password-1")

	assert.Equal(t, serviceBrokers[1].Name, "found-name-2")
	assert.Equal(t, serviceBrokers[1].Guid, "found-guid-2")
	assert.Equal(t, serviceBrokers[1].Url, "http://found.example.com-2")
	assert.Equal(t, serviceBrokers[1].Username, "found-username-2")
	assert.Equal(t, serviceBrokers[1].Password, "found-password-2")
}

func TestFindServiceBrokerByName(t *testing.T) {
	responseBody := `{
  "resources": [
  	{
  	  "metadata": {
  	    "guid":"found-guid"
  	  },
  	  "entity": {
  	    "name": "found-name",
  	    "broker_url": "http://found.example.com",
  	    "auth_username": "found-username",
  	    "auth_password": "found-password"
  	  }
  	}
  ]
}`

	endpoint, status := testapi.CreateCheckableEndpoint(
		"GET",
		"/v2/service_brokers?q=name%3Amy-broker",
		nil,
		testapi.TestResponse{Status: http.StatusOK, Body: responseBody},
	)

	repo, ts := createServiceBrokerRepo(endpoint)
	defer ts.Close()

	foundBroker, apiResponse := repo.FindByName("my-broker")
	expectedBroker := cf.ServiceBroker{
		Name:     "found-name",
		Url:      "http://found.example.com",
		Username: "found-username",
		Password: "found-password",
		Guid:     "found-guid",
	}

	assert.True(t, status.Called())
	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, foundBroker, expectedBroker)
}

func TestFindServiceBrokerByNameWheNotFound(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"GET",
		"/v2/service_brokers?q=name%3Amy-broker",
		nil,
		testapi.TestResponse{Status: http.StatusOK, Body: `{ "resources": [ ] }`},
	)

	repo, ts := createServiceBrokerRepo(endpoint)
	defer ts.Close()

	_, apiResponse := repo.FindByName("my-broker")

	assert.True(t, status.Called())
	assert.True(t, apiResponse.IsNotFound())
	assert.Equal(t, apiResponse.Message, "Service Broker my-broker not found")
}

func TestCreateServiceBroker(t *testing.T) {
	expectedReqBody := `{"name":"foobroker","broker_url":"http://example.com","auth_username":"foouser","auth_password":"password"}`

	endpoint, status := testapi.CreateCheckableEndpoint(
		"POST",
		"/v2/service_brokers",
		testapi.RequestBodyMatcher(expectedReqBody),
		testapi.TestResponse{Status: http.StatusCreated},
	)

	repo, ts := createServiceBrokerRepo(endpoint)
	defer ts.Close()

	serviceBroker := cf.ServiceBroker{
		Name:     "foobroker",
		Url:      "http://example.com",
		Username: "foouser",
		Password: "password",
	}
	apiResponse := repo.Create(serviceBroker)

	assert.True(t, status.Called())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestUpdateServiceBroker(t *testing.T) {
	expectedReqBody := `{"broker_url":"http://update.example.com","auth_username":"update-foouser","auth_password":"update-password"}`

	endpoint, status := testapi.CreateCheckableEndpoint(
		"PUT",
		"/v2/service_brokers/my-guid",
		testapi.RequestBodyMatcher(expectedReqBody),
		testapi.TestResponse{Status: http.StatusOK},
	)

	repo, ts := createServiceBrokerRepo(endpoint)
	defer ts.Close()

	serviceBroker := cf.ServiceBroker{
		Guid:     "my-guid",
		Name:     "foobroker",
		Url:      "http://update.example.com",
		Username: "update-foouser",
		Password: "update-password",
	}
	apiResponse := repo.Update(serviceBroker)

	assert.True(t, status.Called())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestRenameServiceBroker(t *testing.T) {
	expectedReqBody := `{"name":"update-foobroker"}`

	endpoint, status := testapi.CreateCheckableEndpoint(
		"PUT",
		"/v2/service_brokers/my-guid",
		testapi.RequestBodyMatcher(expectedReqBody),
		testapi.TestResponse{Status: http.StatusOK},
	)

	repo, ts := createServiceBrokerRepo(endpoint)
	defer ts.Close()

	serviceBroker := cf.ServiceBroker{
		Guid: "my-guid",
		Name: "update-foobroker",
	}
	apiResponse := repo.Rename(serviceBroker)

	assert.True(t, status.Called())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestDeleteServiceBroker(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"DELETE",
		"/v2/service_brokers/my-guid",
		nil,
		testapi.TestResponse{Status: http.StatusNoContent},
	)

	repo, ts := createServiceBrokerRepo(endpoint)
	defer ts.Close()

	serviceBroker := cf.ServiceBroker{
		Guid: "my-guid",
	}
	apiResponse := repo.Delete(serviceBroker)

	assert.True(t, status.Called())
	assert.True(t, apiResponse.IsSuccessful())
}

func createServiceBrokerRepo(endpoint http.HandlerFunc) (repo ServiceBrokerRepository, ts *httptest.Server) {
	ts = httptest.NewTLSServer(endpoint)

	config := &configuration.Configuration{
		Target:      ts.URL,
		AccessToken: "BEARER my_access_token",
	}

	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerServiceBrokerRepository(config, gateway)
	return
}
