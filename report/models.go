package report

import "github.com/marc47marc47/pandora-go-sdk/base/reqerr"

type ReportToken struct {
	Token string `json:"-"`
}

type UserActivateInput struct {
	ReportToken
}

func (c *UserActivateInput) Validate() error {
	return nil
}

type UserActivateOutput struct {
	Username string `json:"user"`
	Password string `json:"password"`
}

//database related
type CreateDatabaseInput struct {
	ReportToken
	DatabaseName string
	Region       string `json:"region"`
}

func (c *CreateDatabaseInput) Validate() error {
	if c.DatabaseName == "" {
		return reqerr.NewInvalidArgs("Create Database", "database name should not be empty")
	}
	if c.Region == "" {
		return reqerr.NewInvalidArgs("Region", "region should not be empty")
	}
	return nil
}

type ListDatabasesInput struct {
	ReportToken
}

func (c *ListDatabasesInput) Validate() error {

	return nil
}

type ListDatabasesOutput []string

type DeleteDatabaseInput struct {
	ReportToken
	DatabaseName string
}

func (c *DeleteDatabaseInput) Validate() error {
	if c.DatabaseName == "" {
		return reqerr.NewInvalidArgs("Create Database", "database name should not be empty")
	}

	return nil
}

//Table related
type CreateTableInput struct {
	ReportToken
	DatabaseName string
	TableName    string
	CMD          string
}

func (c *CreateTableInput) Validate() error {
	if c.DatabaseName == "" {
		return reqerr.NewInvalidArgs("Create Database", "database name should not be empty")
	}
	if c.TableName == "" {
		return reqerr.NewInvalidArgs("Create Database", "table name should not be empty")
	}
	if c.CMD == "" {
		return reqerr.NewInvalidArgs("Create Database", "create table command should not be empty")

	}
	return nil
}

type UpdateTableInput CreateTableInput

func (c *UpdateTableInput) Validate() error {
	if c.DatabaseName == "" {
		return reqerr.NewInvalidArgs("Create Database", "database name should not be empty")
	}
	if c.TableName == "" {
		return reqerr.NewInvalidArgs("Create Database", "table name should not be empty")
	}
	if c.CMD == "" {
		return reqerr.NewInvalidArgs("Create Database", "create table command should not be empty")

	}
	return nil
}

type ListTablesInput struct {
	ReportToken
	DatabaseName string
}

func (c *ListTablesInput) Validate() error {
	if c.DatabaseName == "" {
		return reqerr.NewInvalidArgs("List Table", "database name should not be empty")
	}

	return nil
}

type GetTableInput struct {
	ReportToken
	DatabaseName string
	TableName    string
}

func (c *GetTableInput) Validate() error {
	if c.DatabaseName == "" {
		return reqerr.NewInvalidArgs("Get Table", "database name should not be empty")
	}
	if c.TableName == "" {
		return reqerr.NewInvalidArgs("Get Database", "table name should not be empty")
	}

	return nil
}

type GetTableOutput []GetTableItem

type GetTableItem struct {
	Field   string      `json:"field"`
	Type    string      `json:"type"`
	Null    string      `json:"null"`
	Key     interface{} `json:"key"`
	Default interface{} `json:"default"`
	Extra   string      `json:"extra"`
}

type ListTablesOutput []string

type DeleteTableInput struct {
	DatabaseName string
	TableName    string
	ReportToken
}

func (c *DeleteTableInput) Validate() error {
	if c.DatabaseName == "" {
		return reqerr.NewInvalidArgs("Delete Table", "database name should not be empty")
	}
	if c.TableName == "" {
		return reqerr.NewInvalidArgs("Create Database", "table name should not be empty")
	}

	return nil
}
