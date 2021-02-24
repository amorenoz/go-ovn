package goovn

type BaseModel struct {
	UUID string `ovs:"_uuid"`
}

type Model interface {
	Table() TableName
}

type TableModelGen func(string) Model
type TableName string
type DBModel map[TableName]TableModelGen
