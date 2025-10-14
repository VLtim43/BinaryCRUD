package dao

import "BinaryCRUD/backend/serialization"

type ItemDAO struct {
	filename string
}

func NewItemDAO(filename string) *ItemDAO {
	return &ItemDAO{filename: filename}
}

func (dao *ItemDAO) Write(text string) error {
	return serialization.AppendEntry(dao.filename, text)
}

func (dao *ItemDAO) Read() ([]serialization.Item, error) {
	return serialization.ReadAllEntries(dao.filename)
}

func (dao *ItemDAO) Print() (string, error) {
	return serialization.PrintBinaryFile(dao.filename)
}
