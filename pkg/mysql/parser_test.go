package mysql

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	s := `
	[mysqld]
	socket="aaa"
	server_id=1
	`
	parser := NewConfigParser()
	if err := parser.Parse(strings.NewReader(s)); err != nil {
		t.Fatal(err.Error())
	}

	mysqldSection, err := parser.GetSection("mysqld")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(mysqldSection.Get("socket"))
}
