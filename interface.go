package wade

var (
	ClientSide bool
)

type (
	Map map[string]interface{}

	// PageControllerFunc is the functiong to be run on the load of a page or page scope
	ControllerFunc func(Context) Map

	Storage interface {
		Get(key string) interface{}
		Set(key string, v interface{})
		Delete(key string)
	}
)
