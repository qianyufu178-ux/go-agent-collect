package errcode

import "fmt"

// ErrorCode represents a unified error with code and message.
type ErrorCode struct {
	Code    int
	Message string
}

func (e *ErrorCode) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Wrap creates a new ErrorCode with additional context.
func (e *ErrorCode) Wrap(context string) *ErrorCode {
	return &ErrorCode{
		Code:    e.Code,
		Message: fmt.Sprintf("%s: %s", context, e.Message),
	}
}

// Predefined error codes.
// Range: 1000-1999 for Edge errors, 2000-2999 for Controller errors.
var (
	// Common errors (1000-1099)
	OK                 = &ErrorCode{Code: 0, Message: "success"}
	ErrInternal        = &ErrorCode{Code: 1000, Message: "internal error"}
	ErrConfig          = &ErrorCode{Code: 1001, Message: "configuration error"}
	ErrNotImplemented  = &ErrorCode{Code: 1002, Message: "not implemented"}
	ErrTimeout         = &ErrorCode{Code: 1003, Message: "operation timeout"}
	ErrInvalidParam    = &ErrorCode{Code: 1004, Message: "invalid parameter"}

	// Edge errors (1100-1199)
	ErrProxyStart      = &ErrorCode{Code: 1100, Message: "proxy start failed"}
	ErrProxyListen     = &ErrorCode{Code: 1101, Message: "proxy listen failed"}
	ErrRateLimit       = &ErrorCode{Code: 1102, Message: "rate limit exceeded"}
	ErrCollectorStart  = &ErrorCode{Code: 1103, Message: "collector start failed"}
	ErrCacheWrite      = &ErrorCode{Code: 1104, Message: "cache write failed"}
	ErrCacheRead       = &ErrorCode{Code: 1105, Message: "cache read failed"}
	ErrCacheFull       = &ErrorCode{Code: 1106, Message: "cache storage full"}
	ErrNetworkDown     = &ErrorCode{Code: 1107, Message: "network unavailable"}
	ErrGRPCConn        = &ErrorCode{Code: 1108, Message: "gRPC connection failed"}
	ErrHeartbeat       = &ErrorCode{Code: 1109, Message: "heartbeat failed"}

	// Controller errors (2000-2099)
	ErrNodeNotFound    = &ErrorCode{Code: 2000, Message: "node not found"}
	ErrNodeOffline     = &ErrorCode{Code: 2001, Message: "node is offline"}
	ErrStrategyInvalid = &ErrorCode{Code: 2002, Message: "invalid strategy"}
	ErrStrategyPush    = &ErrorCode{Code: 2003, Message: "strategy push failed"}
	ErrDBNotReady      = &ErrorCode{Code: 2004, Message: "database not ready"}
	ErrDBWrite         = &ErrorCode{Code: 2005, Message: "database write failed"}
	ErrDBQuery         = &ErrorCode{Code: 2006, Message: "database query failed"}
	ErrAuthFailed      = &ErrorCode{Code: 2007, Message: "authentication failed"}
)
