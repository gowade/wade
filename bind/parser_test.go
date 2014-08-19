package bind

import (
	"reflect"
	"testing"
)

type TestUser struct {
	Test string
	Data struct {
		Username string
		Password string
	}
}

func TestParser(t *testing.T) {
	model := new(TestUser)
	model.Data.Username = "Hai"
	model.Data.Password = "Pk"
	model.Test = "T"

	helpers := map[string]interface{}{
		"addInt": func(a, b int) int {
			return a + b
		},
		"addFloat": func(a, b float32) float32 {
			return a + b
		},
		"fooAdd": func(str string) string {
			return "foo" + str
		},
	}

	hst := helpersSymbolTable(helpers)
	dhst := helpersSymbolTable(defaultHelpers())
	bs := &bindScope{&scope{[]symbolTable{dhst, hst, modelSymbolTable{reflect.ValueOf(model)}}}}
	tests := map[string]interface{}{
		"Test":                                                   "T",
		"Data.Username":                                          "Hai",
		"toUpper(Data.Username)":                                 "HAI",
		"concat(Data.Username, Data.Password)":                   "HaiPk",
		"concat(toUpper(Data.Username), toLower(Data.Password))": "HAIpk",
		"addInt(1, 2)":       3,
		"addFloat(1.0, 2.0)": float32(3),
		"fooAdd('bar-,')":    "foobar-,",
	}

	for bstr, result := range tests {
		_, _, v, err := bs.evaluate(bstr)
		if err != nil {
			t.Fatal(err)
		}
		switch v.(type) {
		case string, int, float32:
			if v != result {
				t.Errorf("Expected %v, got %v.", result, v)
			}
		}
	}

	errtests := []string{
		`fooAdd('bar)`,
		`kdf*`,
		`toUpper(Data.Username.)`,
		`addInt(1a)`,
		`addInt(''')`,
		`addInt('*,')`,
	}
	for _, et := range errtests {
		_, _, _, err := bs.evaluate(et)
		if err == nil {
			t.Errorf("Expected an error, no error is returned.")
		} else {
			//t.Logf("Log: got parse error: %s\n", err)
		}
	}
}
