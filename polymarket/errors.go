package polymarket

import "fmt"

type ApiError struct {
	Message string
	Status  int
	Data    any
}

func (e *ApiError) Error() string {
	if e == nil {
		return ""
	}
	if e.Status > 0 {
		return fmt.Sprintf("%s (status=%d)", e.Message, e.Status)
	}
	return e.Message
}

var (
	ErrL1AuthUnavailable    = fmt.Errorf("Signer is needed to interact with this endpoint!")
	ErrL2AuthNotAvailable   = fmt.Errorf("API Credentials are needed to interact with this endpoint!")
	ErrBuilderAuthRequired  = fmt.Errorf("Builder API Credentials needed to interact with this endpoint!")
	ErrBuilderAuthFailed    = fmt.Errorf("Builder key auth failed!")
	ErrOrderSignerMissing   = fmt.Errorf("order signer is required for order creation endpoints")
	ErrBuilderNotConfigured = fmt.Errorf("builder configuration is required")
)
