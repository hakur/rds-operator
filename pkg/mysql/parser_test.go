package mysql

import (
	"fmt"
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

func TestParserMergeSection(t *testing.T) {
	s := `
	[mysqld]
	socket="aaa"
	server_id=1
	`
	parser := NewConfigParser()
	if err := parser.Parse(strings.NewReader(s)); err != nil {
		t.Fatal(err.Error())
	}

	override := NewConfigSection("mysqld")

	override.Set("socket", "bbb")
	override.Set("scoket", "ddd")
	override.Set("xxx", "ccc")

	if err := parser.MergeSection(override); err != nil {
		t.Fatal(err.Error())
	}
	mysqldSection, err := parser.GetSection("mysqld")
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(mysqldSection.Data)
	t.Log(mysqldSection.Get("socket"))
}
