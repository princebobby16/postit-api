package crud

import (
	"database/sql"
	"strconv"
	"strings"
)

type Service interface {
	CreateResource() int
	FetchResource() []struct{}
	UpdateResource() int
	DeleteResource() int
}

type Object struct {
	Table        string
	Fields       []string
	Schema       string
	BbConnection *sql.DB
}

func (o Object) CreateResource(value interface{}) (int64, error) {
	tableProperties := o.Schema + "." + o.Table

	// build primary values
	fields := ""
	for _, val := range o.Fields {
		fields += val + ", "
	}

	// build the secondary values
	dollarSign := ""
	for i := 0; i < len(o.Fields); i++ {
		dollarSign += "$" + strconv.Itoa(i+1) + ", "
	}

	//formatted strings
	fieldsFormatted := strings.TrimRight(fields, ", ")
	dollarSignFormatted := strings.TrimRight(dollarSign, ", ")

	query := "INSERT INTO " + tableProperties + "(" + fieldsFormatted + ")" + " VALUES (" + dollarSignFormatted + ")"
	a, err := o.BbConnection.Exec(query)
	if err != nil {
		return 1, err
	}

	i, err := a.LastInsertId()
	if err != nil {
		return 2, err
	}

	return i, nil
}

func (o Object) FetchResource() []struct{} {
	panic("implement me")
}

func (o Object) UpdateResource() int {
	panic("implement me")
}

func (o Object) DeleteResource() int {
	panic("implement me")
}
