package client_test

import (
	"github.com/go-chassis/go-sc-client"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/go-chassis/go-chassis/core/lager"
	"github.com/go-chassis/paas-lager"
	"github.com/go-mesh/openlogging"
	"os"
	"time"
)

func init() {
	log.Init(log.Config{
		LoggerLevel:   "DEBUG",
		EnableRsyslog: false,
		LogFormatText: true,
		Writers:       []string{"stdout"},
	})
	l := log.NewLogger("test")
	openlogging.SetLogger(l)
}
func TestLoadbalance(t *testing.T) {

	t.Log("Testing Round robin function")
	var sArr []string

	sArr = append(sArr, "s1")
	sArr = append(sArr, "s2")

	next := client.RoundRobin(sArr)
	_, err := next()
	assert.NoError(t, err)
}

func TestLoadbalanceEmpty(t *testing.T) {
	t.Log("Testing Round robin with empty endpoint arrays")
	var sArrEmpty []string

	next := client.RoundRobin(sArrEmpty)
	_, err := next()
	assert.Error(t, err)

}

func TestClientInitializeHttpErr(t *testing.T) {
	t.Log("Testing for HTTPDo function with errors")

	hostname, err := os.Hostname()
	if err != nil {
		openlogging.GetLogger().Error("Get hostname failed.")
		return
	}
	microServiceInstance := &client.MicroServiceInstance{
		Endpoints: []string{"rest://127.0.0.1:3000"},
		HostName:  hostname,
		Status:    client.MSInstanceUP,
	}

	registryClient := &client.RegistryClient{}

	err = registryClient.Initialize(
		client.Options{
			Addrs: []string{"127.0.0.1:30100"},
		})
	assert.NoError(t, err)

	err = registryClient.SyncEndpoints()
	assert.NoError(t, err)

	httpHeader := registryClient.GetDefaultHeaders()
	assert.NotEmpty(t, httpHeader)

	resp, err := registryClient.HTTPDo("GET", "fakeRawUrl", httpHeader, []byte("fakeBody"))
	assert.Empty(t, resp)
	assert.Error(t, err)

	MSList, err := registryClient.GetAllMicroServices()
	assert.NotEmpty(t, MSList)
	assert.NoError(t, err)

	f1 := func(*client.MicroServiceInstanceChangedEvent) {}
	err = registryClient.WatchMicroService(MSList[0].ServiceID, f1)
	assert.NoError(t, err)

	var ms = new(client.MicroService)
	var msdepreq = new(client.MircroServiceDependencyRequest)
	var msdepArr []*client.MicroServiceDependency
	var msdep1 = new(client.MicroServiceDependency)
	var msdep2 = new(client.MicroServiceDependency)
	var dep = new(client.DependencyMicroService)
	var m = make(map[string]string)

	m["abc"] = "abc"
	m["def"] = "def"

	dep.AppID = "appid"

	msdep1.Consumer = dep
	msdep2.Consumer = dep

	msdepArr = append(msdepArr, msdep1)
	msdepArr = append(msdepArr, msdep2)

	ms.AppID = MSList[0].AppID
	ms.ServiceName = MSList[0].ServiceName
	ms.Version = MSList[0].Version
	ms.Environment = MSList[0].Environment
	ms.Properties = m

	msdepreq.Dependencies = msdepArr
	s1, err := registryClient.RegisterMicroServiceInstance(microServiceInstance)
	assert.Empty(t, s1)
	assert.Error(t, err)

	s1, err = registryClient.RegisterMicroServiceInstance(nil)
	assert.Empty(t, s1)
	assert.Error(t, err)

	msArr, err := registryClient.GetMicroServiceInstances("fakeConsumerID", "fakeProviderID")
	assert.Empty(t, msArr)
	assert.Error(t, err)

	msArr, err = registryClient.Health()
	assert.NotEmpty(t, msArr)
	assert.NoError(t, err)

	b, err := registryClient.UpdateMicroServiceProperties(MSList[0].ServiceID, ms)
	assert.Equal(t, true, b)
	assert.NoError(t, err)

	f1 = func(*client.MicroServiceInstanceChangedEvent) {}
	err = registryClient.WatchMicroService(MSList[0].ServiceID, f1)
	assert.NoError(t, err)

	f1 = func(*client.MicroServiceInstanceChangedEvent) {}
	err = registryClient.WatchMicroService("", f1)
	assert.Error(t, err)

	f1 = func(*client.MicroServiceInstanceChangedEvent) {}
	err = registryClient.WatchMicroService(MSList[0].ServiceID, nil)
	assert.NoError(t, err)

	str, err := registryClient.RegisterService(ms)
	assert.NotEmpty(t, str)
	assert.NoError(t, err)

	str, err = registryClient.RegisterService(nil)
	assert.Empty(t, str)
	assert.Error(t, err)

	ms1, err := registryClient.GetProviders("fakeconsumer")
	assert.Empty(t, ms1)
	assert.Error(t, err)

	err = registryClient.AddDependencies(msdepreq)
	assert.Error(t, err)

	err = registryClient.AddDependencies(nil)
	assert.Error(t, err)

	err = registryClient.AddSchemas(MSList[0].ServiceID, "schema", "schema")
	assert.NoError(t, err)

	getms1, err := registryClient.GetMicroService(MSList[0].ServiceID)
	assert.NotEmpty(t, getms1)
	assert.NoError(t, err)

	getms2, err := registryClient.FindMicroServiceInstances("abcd", MSList[0].AppID, MSList[0].ServiceName, MSList[0].Version)
	assert.Empty(t, getms2)
	assert.Error(t, err)

	getmsstr, err := registryClient.GetMicroServiceID(MSList[0].AppID, MSList[0].ServiceName, MSList[0].Version, MSList[0].Environment)
	assert.NotEmpty(t, getmsstr)
	assert.NoError(t, err)

	getmsstr, err = registryClient.GetMicroServiceID(MSList[0].AppID, "Server112", MSList[0].Version, "")
	assert.Empty(t, getmsstr)
	//assert.Error(t, err)

	ms.Properties = nil
	b, err = registryClient.UpdateMicroServiceProperties(MSList[0].ServiceID, ms)
	assert.Equal(t, false, b)
	assert.Error(t, err)

	err = registryClient.AddSchemas("", "schema", "schema")
	assert.Error(t, err)

	b, err = registryClient.Heartbeat(MSList[0].ServiceID, "")
	assert.Equal(t, false, b)
	assert.Error(t, err)

	b, err = registryClient.UpdateMicroServiceInstanceStatus(MSList[0].ServiceID, "", MSList[0].Status)
	assert.Equal(t, false, b)
	assert.Error(t, err)

	b, err = registryClient.UnregisterMicroService("")
	assert.Equal(t, false, b)
	assert.Error(t, err)
	services, err := registryClient.GetAllResources("instances")
	assert.NotZero(t, len(services))
	assert.NoError(t, err)
	err = registryClient.Close()
	assert.NoError(t, err)

}
func TestRegistryClient_FindMicroServiceInstances(t *testing.T) {
	lager.Initialize("", "DEBUG", "",
		"size", true, 1, 10, 7)

	hostname, err := os.Hostname()
	if err != nil {
		openlogging.GetLogger().Error("Get hostname failed.")
		return
	}
	ms := &client.MicroService{
		ServiceName: "Server",
		AppID:       "default",
		Version:     "0.0.1",
	}
	var sid string
	registryClient := &client.RegistryClient{}

	err = registryClient.Initialize(
		client.Options{
			Addrs: []string{"127.0.0.1:30100"},
		})
	assert.NoError(t, err)
	sid, err = registryClient.RegisterService(ms)
	if err == client.ErrMicroServiceExists {
		sid, err = registryClient.GetMicroServiceID("default", "Server", "0.0.1", "")
		assert.NoError(t, err)
		assert.NotNil(t, sid)
	}

	microServiceInstance := &client.MicroServiceInstance{
		ServiceID: sid,
		Endpoints: []string{"rest://127.0.0.1:3000"},
		HostName:  hostname,
		Status:    client.MSInstanceUP,
	}

	iid, err := registryClient.RegisterMicroServiceInstance(microServiceInstance)
	assert.NotNil(t, iid)
	_, err = registryClient.FindMicroServiceInstances(sid, "default", "Server", "0.0.1")
	assert.NoError(t, err)

	t.Log("find again, should get ErrNotModified")
	_, err = registryClient.FindMicroServiceInstances(sid, "default", "Server", "0.0.1")
	assert.Equal(t, client.ErrNotModified, err)

	t.Log("find again without revision, should get nil error")
	_, err = registryClient.FindMicroServiceInstances(sid, "default", "Server", "0.0.1", client.WithoutRevision())
	assert.NoError(t, err)

	t.Log("register new and find")
	microServiceInstance2 := &client.MicroServiceInstance{
		ServiceID: sid,
		Endpoints: []string{"rest://127.0.0.1:3001"},
		HostName:  hostname + "1",
		Status:    client.MSInstanceUP,
	}
	iid, err = registryClient.RegisterMicroServiceInstance(microServiceInstance2)
	time.Sleep(3 * time.Second)
	_, err = registryClient.FindMicroServiceInstances(sid, "default", "Server", "0.0.1")
	assert.NoError(t, err)

	t.Log("after reset")
	registryClient.ResetRevision()
	_, err = registryClient.FindMicroServiceInstances(sid, "default", "Server", "0.0.1")
	assert.NoError(t, err)
	_, err = registryClient.FindMicroServiceInstances(sid, "default", "Server", "0.0.1")
	assert.Equal(t, client.ErrNotModified, err)

	_, err = registryClient.FindMicroServiceInstances(sid, "appIDNotExists", "ServerNotExists", "0.0.1")
	assert.Equal(t, client.ErrMicroServiceNotExists, err)

}
func TestRegistryClient_GetDefaultHeaders(t *testing.T) {
	registryClient := &client.RegistryClient{}

	err := registryClient.Initialize(
		client.Options{
			Addrs:        []string{"127.0.0.1:30100"},
			ConfigTenant: "go-sc-tenant",
		})
	assert.Nil(t, err)

	header := registryClient.GetDefaultHeaders()
	tenant := header.Get(client.TenantHeader)
	assert.Equal(t, tenant, "go-sc-tenant")
}
