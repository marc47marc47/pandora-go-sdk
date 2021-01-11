package report

import (
	"github.com/marc47marc47/pandora-go-sdk/base"
)

type ReportAPI interface {
	ActivateUser(*UserActivateInput) (*UserActivateOutput, error)

	CreateDatabase(*CreateDatabaseInput) error

	ListDatabases(*ListDatabasesInput) (*ListDatabasesOutput, error)

	DeleteDatabase(*DeleteDatabaseInput) error

	CreateTable(*CreateTableInput) error

	UpdateTable(*UpdateTableInput) error

	ListTables(*ListTablesInput) (*ListTablesOutput, error)

	GetTable(*GetTableInput) (*GetTableOutput, error)

	DeleteTable(*DeleteTableInput) error

	MakeToken(*base.TokenDesc) (string, error)
}
