package core

import (
	"reflect"
	"testing"
	"time"

	. "github.com/phaikawl/wade/scope"
)

type TestUser struct {
	Test string
	Data struct {
		Username string
		Password string
	}
}

func (tu *TestUser) TestConcat(s string) {
	tu.Test += s
}

func TestParser(t *testing.T) {
	model := new(TestUser)
	model.Data.Username = "Hai"
	model.Data.Password = "Pk"
	model.Test = "Nt"

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
		"model": func() *TestUser {
			return model
		},
	}

	hst := NewHelpersSymbolTable(helpers)
	dhst := NewHelpersSymbolTable(defaultHelpers())
	bs := &bindScope{NewScope([]SymbolTable{dhst, hst, NewModelSymbolTable(model)})}
	tests := map[string]interface{}{
		"$Test":                                                     "Nt",
		"$Data.Username":                                            "Hai",
		"toUpper($Data.Username)":                                   "HAI",
		"strJoin($Data.Username, $Data.Password)":                   "HaiPk",
		"strJoin($Data.Username, 'Pk|')":                            "HaiPk|",
		"strJoin(toUpper($Data.Username), toLower($Data.Password))": "HAIpk",
		"addInt(-1, 2)":                                             1,
		"addFloat(-1.0, 2.0)":                                       float32(1.0),
		"fooAdd('bar*|-,')":                                         "foobar*|-,",
		"strJoin($model().Test, 'a')":                               "Nta",
	}

	for bstr, result := range tests {
		_, _, v, err := bs.evaluate(bstr)
		if err != nil {
			t.Fatalf(`Error {%v"} on bind string "%v".`, err.Error(), bstr)
		}
		switch v.(type) {
		case string, int, float32:
			if v != result {
				t.Errorf("Expected %v, got %v.", result, v)
			}
		}
	}

	_, _, v, err := bs.evaluate("@TestConcat($Data.Username)")
	if err != nil {
		t.Fatal(err)
	}

	reflect.ValueOf(v).Call([]reflect.Value{})
	go func() {
		time.Sleep(200)
		if model.Test != "NtHai" {
			t.Errorf("Expected %v, got %v.", "NtHai", model.Test)
		}
	}()

	errtests := []string{
		`fooAdd('bar'`,
		`kdf*`,
		`toUpper(Data.Username.)`,
		`addInt(1, 1a)`,
		`fooAdd(''')`,
		`addInt(1, '*,')`,
	}
	for _, et := range errtests {
		_, _, _, err := bs.evaluate(et)
		if err == nil {
			t.Errorf("Bind string `%v`, Expected an error, no error is returned.", et)
		} else {
			//t.Logf("Log: got parse error: %s\n", err)
		}
	}
}
