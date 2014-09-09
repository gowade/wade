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

	hst := newHelpersSymbolTable(helpers)
	dhst := newHelpersSymbolTable(defaultHelpers())
	bs := &bindScope{&scope{[]symbolTable{dhst, hst, modelSymbolTable{reflect.ValueOf(model)}}}}
	tests := map[string]interface{}{
		"Test":                                                            "T",
		"Data.Username":                                                   "Hai",
		"Data.Username | toUpper(@1)":                                     "HAI",
		"Data.Username, Data.Password | concat(@1, @2)":                   "HaiPk",
		"Data.Username | concat(@1, 'Pk|')":                               "HaiPk|",
		"Data.Username, Data.Password | concat(toUpper(@1), toLower(@2))": "HAIpk",
		"| addInt(-1, 2)":                                                 1,
		"| addFloat(-1.0, 2.0)":                                           float32(1.0),
		"| fooAdd('bar*|-,')":                                             "foobar*|-,",
	}

	for bstr, result := range tests {
		_, _, _, v, err := bs.evaluate(bstr)
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
		`fooAdd('bar'`,
		`fooAdd('bar')`,
		`kdf*`,
		`| toUpper(Data.Username.)`,
		`| addInt(1, 1a)`,
		`| fooAdd(''')`,
		`| addInt(1, '*,')`,
	}
	for _, et := range errtests {
		_, _, _, _, err := bs.evaluate(et)
		if err == nil {
			t.Errorf("Expected an error, no error is returned.")
		} else {
			//t.Logf("Log: got parse error: %s\n", err)
		}
	}
}
