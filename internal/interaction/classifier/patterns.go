package classifier

// Exploit type constants
const (
	ExploitLog4shell       = "log4shell"
	ExploitFastjson        = "fastjson"
	ExploitXXE             = "xxe"
	ExploitSSRF            = "ssrf"
	ExploitSQLi            = "sqli"
	ExploitRCE             = "rce"
	ExploitDeserialization = "deserialization"
	ExploitLDAP            = "ldap"
	ExploitXSS             = "xss"
)

// Confidence level constants
const (
	ConfidenceHigh   = "high"
	ConfidenceMedium = "medium"
	ConfidenceLow    = "low"
)
