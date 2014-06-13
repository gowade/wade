package lib

const (
	InvalidCharacters string = "Invalid characters."
)

type FormResp struct {
	Ok     bool
	Errors map[string]string
}
