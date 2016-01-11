package lars

// type HandlerFunc func(Context)

const (
	blank = ""
)

// HandlerFunc type defines a single handler
type HandlerFunc func(Context)

// HandlersChain type defines an array of Handlers
type HandlersChain []HandlerFunc

// LARS is the main routing instance
type LARS struct {
	// mostParams used to keep track of the most amount of
	// params in any URL and this will set the default capacity
	// of each context Params
	mostParams uint8
}

// New Creates and returns a new TRI instance
func New() *LARS {

	t := &LARS{
		mostParams: 0,
	}

	return t
}
