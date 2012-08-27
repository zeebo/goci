package notifications

//Config stores the configuration for sending notifications
var Config struct {
	Username string //the username of the XMPP sender
	Password string //the password of the XMPP sender
	Domain   string //the domain of the XMPP sender
}
