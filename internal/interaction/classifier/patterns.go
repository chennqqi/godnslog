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

// Pattern constants for matching
const (
	JNDIPattern         = "${jndi:"
	FastjsonTypePattern = `{"@type":"`
	PathXXE             = "/xxe"
	PathXML             = "/xml"
	PathCmd             = "/cmd"
	PathExec            = "/exec"
	DNSSQLPattern       = "sql"
	DNSCMDPattern       = "cmd"
	DNSExecPattern      = "exec"
	DNSJNDIPattern      = "jndi"
	DNSLDAPPattern      = "ldap"
	DNSXXEPattern       = "xxe"
	DNSRFIPattern       = "rfi"
	DNSDeserPattern     = "object"
	BodyCmdPattern      = "whoami"
	BodyExecPattern     = "exec"
)
