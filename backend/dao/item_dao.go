package dao

import "BinaryCRUD/backend/serialization"

type ItemDAO struct {
	filename string
}

func NewItemDAO(filename string) *ItemDAO {
	return &ItemDAO{filename: filename}
}

func (dao *ItemDAO) Write(text string) error {
	return serialization.WriteHex(dao.filename, text)
}

func (dao *ItemDAO) Read() (string, error) {
	return serialization.ReadHex(dao.filename)
}
