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
	{"", "localhost:30004"},
	{"", "localhost:30005"}}

var MachinesSSH = [...]Machine{
	{"lab102-195.cps.unizar.es", "155.210.154.195:30000"},
	{"lab102-196.cps.unizar.es", "155.210.154.196:30000"},
	{"lab102-197.cps.unizar.es", "155.210.154.197:30000"},
	{"lab102-198.cps.unizar.es", "155.210.154.198:30000"},
	{"lab102-199.cps.unizar.es", "155.210.154.199:30000"},
	{"lab102-200.cps.unizar.es", "155.210.154.200:30000"}}
