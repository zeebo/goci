package rpc

//Config represents a json struct that is read in by the Builder. It can describe
//things like notifcations or other configuration data for the test. It is loaded
//from files named `.goci` in the package directory. The file is loaded by descending
//the package heiarchy loading in the config and overwriting values as it goes,
//so that config files lower in the path inherit the values from higher in the
//path.
type Config struct {
	//omitempty is used so that values set previously don't get overwritten by
	//empty values in the next read.
	NotifyJabber string `json:",omitempty"` // a jabber address for an XMPP message
	NotifyOn     string `json:",omitempty"` // one of: `pass`, `fail`, `error`, `wontbuild`, `problem`, `always`, `change`
	NotifyURL    string `json:",omitempty"` // a URL that will be posted with the result data
}
