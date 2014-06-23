package wade

import "testing"

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
	b := newBindEngine(nil)
	tests := map[string]string{
		"Test":                                                   "T",
		"Data.Username":                                          "Hai",
		"toUpper(Data.Username)":                                 "HAI",
		"concat(Data.Username, Data.Password)":                   "HaiPk",
		"concat(toUpper(Data.Username), toLower(Data.Password))": "HAIpk",
	}

	for bs, result := range tests {
		_, _, v := b.evaluateBindString(bs, model)
		if v.(string) != result {
			t.Errorf("Expected %v, got %v.", result, v.(string))
		}
	}
}
