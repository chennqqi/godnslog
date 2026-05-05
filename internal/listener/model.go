package listener

import (
	"github.com/chennqqi/godnslog/internal/models"
)

// Re-export types from models package for backward compatibility
type Protocol = models.Protocol
type Listener = models.Listener
type ListenerInteraction = models.ListenerInteraction
type Metadata = models.Metadata
type SMTPMessage = models.SMTPMessage
type LDAPQuery = models.LDAPQuery
type SMBRequest = models.SMBRequest
type FTPCommand = models.FTPCommand
type ListenerConfig = models.ListenerConfig
type ListenerListResponse = models.ListenerListResponse

// Re-export constants
const (
	ProtocolSMTP = models.ProtocolSMTP
	ProtocolLDAP = models.ProtocolLDAP
	ProtocolSMB  = models.ProtocolSMB
	ProtocolFTP  = models.ProtocolFTP
)
