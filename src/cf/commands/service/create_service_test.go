package service_test

import (
	"cf"
	"cf/api"
	. "cf/commands/service"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestCreateService(t *testing.T) {
	serviceOfferings := []cf.ServiceOffering{
		cf.ServiceOffering{Label: "cleardb", Plans: []cf.ServicePlan{
			cf.ServicePlan{Name: "spark", Guid: "cleardb-spark-guid"},
		}},
		cf.ServiceOffering{Label: "postgres"},
	}
	serviceRepo := &testapi.FakeServiceRepo{ServiceOfferings: serviceOfferings}
	fakeUI := callCreateService(
		[]string{"cleardb", "spark", "my-cleardb-service"},
		[]string{},
		serviceRepo,
	)

	assert.Contains(t, fakeUI.Outputs[0], "Creating service")
	assert.Contains(t, fakeUI.Outputs[0], "my-cleardb-service")
	assert.Equal(t, serviceRepo.CreateServiceInstanceName, "my-cleardb-service")
	assert.Equal(t, serviceRepo.CreateServiceInstancePlan, cf.ServicePlan{Name: "spark", Guid: "cleardb-spark-guid"})
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestCreateServiceWhenServiceAlreadyExists(t *testing.T) {
	serviceOfferings := []cf.ServiceOffering{
		cf.ServiceOffering{Label: "cleardb", Plans: []cf.ServicePlan{
			cf.ServicePlan{Name: "spark", Guid: "cleardb-spark-guid"},
		}},
		cf.ServiceOffering{Label: "postgres"},
	}
	serviceRepo := &testapi.FakeServiceRepo{ServiceOfferings: serviceOfferings, CreateServiceAlreadyExists: true}
	fakeUI := callCreateService(
		[]string{"cleardb", "spark", "my-cleardb-service"},
		[]string{},
		serviceRepo,
	)

	assert.Contains(t, fakeUI.Outputs[0], "Creating service")
	assert.Contains(t, fakeUI.Outputs[0], "my-cleardb-service")
	assert.Equal(t, serviceRepo.CreateServiceInstanceName, "my-cleardb-service")
	assert.Equal(t, serviceRepo.CreateServiceInstancePlan, cf.ServicePlan{Name: "spark", Guid: "cleardb-spark-guid"})
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "my-cleardb-service")
	assert.Contains(t, fakeUI.Outputs[2], "already exists")
}

func callCreateService(args []string, inputs []string, serviceRepo api.ServiceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{Inputs: inputs}
	ctxt := testcmd.NewContext("create-service", args)
	cmd := NewCreateService(fakeUI, serviceRepo)
	reqFactory := &testreq.FakeReqFactory{}

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
