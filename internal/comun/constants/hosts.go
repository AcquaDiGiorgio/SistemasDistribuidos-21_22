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

//
//{"lab102-198.cps.unizar.es", "155.210.154.198:29420"},
//{"lab102-199.cps.unizar.es", "155.210.154.199:29420"}
//
