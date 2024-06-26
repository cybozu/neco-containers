package main

type Machine struct {
	Hostname  string
	BmcIPadr  string
	NodeIPadr string
	SerialNo  string
}

type Machines struct {
	machine []Machine
}
