package main_test

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"
)

type GenTestSuite struct {
	suite.Suite
}

func (s *GenTestSuite) test(filename string) {
	cmd := exec.Command("fuel", filename)
	op, err := cmd.CombinedOutput()
	fmt.Println(string(op))
	s.NoError(err)
}

func (s *GenTestSuite) TestCodegen() {
	gg := exec.Command("go", "install")
	op, err := gg.CombinedOutput()
	fmt.Println(string(op))
	s.NoError(err)

	s.test("htmltest/test.html")

	gb := exec.Command("go", "build")
	gb.Dir = "htmltest"
	op, err = gb.CombinedOutput()
	fmt.Println(string(op))
	s.NoError(err)
}

func TestCompile(t *testing.T) {
	suite.Run(t, new(GenTestSuite))
}
