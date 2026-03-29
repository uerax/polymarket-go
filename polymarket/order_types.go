package polymarket

type SignatureType int

const (
	SignatureTypeEOA SignatureType = iota
	SignatureTypePolyProxy
	SignatureTypePolyGnosisSafe
)

type BuilderHeaderProvider interface {
	GenerateBuilderHeaders(method string, path string, body string) (map[string]string, error)
	IsValid() bool
}
