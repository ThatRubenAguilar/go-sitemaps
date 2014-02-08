package sitemap/netrule

import (
	"net/http"
)

type NetAccessRule interface {
	// CanAccess should determine whether it is ready to access a website or not
	CanAccess() bool
	// Accessed is called after every net access
	Accessed()
}