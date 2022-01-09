package constants

const USERS = 3

type Machine struct {
	Host string
	Ip   string
}

var MachinesLocal = [...]Machine{
	{"", "localhost:30000"},
	{"", "localhost:30001"},
	{"", "localhost:30002"},
	{"", "localhost:30003"},
	{"", "localhost:30004"}}

var MachinesSSH = [...]Machine{
	{"lab102-195.cps.unizar.es", "155.210.154.195:29420"},
	{"lab102-196.cps.unizar.es", "155.210.154.196:29420"},
	{"lab102-197.cps.unizar.es", "155.210.154.197:29420"}}

var MachinesKubernetesPods = [...]Machine{
	{"", "nr1.conexion.default.svc.cluster.local:7000"},
	{"", "nr2.conexion.default.svc.cluster.local:7000"},
	{"", "nr3.conexion.default.svc.cluster.local:7000"}}

var MachinesKubernetesDeployment = [...]Machine{
	{"", "192.168.0.1:7000"},
	{"", "192.168.0.2:7000"},
	{"", "192.168.0.3:7000"}}
