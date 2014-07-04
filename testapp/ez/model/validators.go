package model

import "github.com/pengux/check"

func UsernameValidator() check.Validator {
	return check.Composite{
		check.NonEmpty{},
		check.Regex{`^[a-zA-Z0-9]+$`},
		check.MinChar{3},
	}
}

func PasswordValidator() check.Validator {
	return check.Composite{
		check.NonEmpty{},
		check.MinChar{6},
	}
}

func UsernamePasswordValidator() check.Struct {
	return check.Struct{
		"Username": UsernameValidator(),
		"Password": PasswordValidator(),
	}
}
