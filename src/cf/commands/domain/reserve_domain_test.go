package domain_test

import (
	"cf"
	. "cf/commands/domain"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestReserveDomainRequirements(t *testing.T) {
	domainRepo := &testapi.FakeDomainRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	callReserveDomain([]string{"my-org", "example.com"}, reqFactory, domainRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.OrganizationName, "my-org")

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}

	callReserveDomain([]string{"my-org", "example.com"}, reqFactory, domainRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestReserveDomainFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	domainRepo := &testapi.FakeDomainRepository{}
	ui := callReserveDomain([]string{""}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callReserveDomain([]string{"org1"}, reqFactory, domainRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callReserveDomain([]string{"org1", "example.com"}, reqFactory, domainRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestReserveDomain(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Organization: cf.Organization{Name: "myOrg", Guid: "myOrg-guid"}}
	domainRepo := &testapi.FakeDomainRepository{}
	fakeUI := callReserveDomain([]string{"myOrg", "example.com"}, reqFactory, domainRepo)

	assert.Equal(t, domainRepo.ReserveDomainDomainToCreate.Name, "example.com")
	assert.Equal(t, domainRepo.ReserveDomainOwningOrg.Name, "myOrg")
	assert.Contains(t, fakeUI.Outputs[0], "example.com")
	assert.Contains(t, fakeUI.Outputs[0], "myOrg")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func callReserveDomain(args []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("reserve-domain", args)
	cmd := NewReserveDomain(fakeUI, domainRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
