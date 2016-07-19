package main

type Model struct {
	Name   string
	Fields []*ModelField
}

type ModelField struct {
	Name     string
	JSONName string
	Type     string
}
