package sqlite

import "github.com/sejo412/gophkeeper/internal/models"

var actions = map[models.RecordType]map[action]query{
	models.RecordPassword: {
		actionCreate: {
			query: queryWithTable("INSERT INTO %s(uid, login, password, meta) VALUES (?, ?, ?, ?)", tablePasswords),
		},
		actionRead: {
			query: queryWithTable("SELECT id, login, password, meta FROM %s WHERE id = ? AND uid = ?", tablePasswords),
		},
		actionUpdate: {
			query: queryWithTable("UPDATE %s SET login = ?, password = ?, meta = ? WHERE id = ? AND uid = ?", tablePasswords),
		},
		actionDelete: {
			query: queryWithTable("DELETE FROM %s WHERE id = ? AND uid = ?", tablePasswords),
		},
		actionList: {
			query: queryWithTable("SELECT id, meta FROM %s WHERE uid = ?", tablePasswords),
		},
	},
	models.RecordText: {
		actionCreate: {
			query: queryWithTable("INSERT INTO %s(uid, text, meta) VALUES (?, ?, ?)", tableTexts),
		},
		actionRead: {
			query: queryWithTable("SELECT id, text, meta FROM %s WHERE id = ? AND uid = ?", tableTexts),
		},
		actionUpdate: {
			query: queryWithTable("UPDATE %s SET text = ?, meta = ? WHERE id = ? AND uid = ?", tableTexts),
		},
		actionDelete: {
			query: queryWithTable("DELETE FROM %s WHERE id = ? AND uid = ?", tableTexts),
		},
		actionList: {
			query: queryWithTable("SELECT id, meta FROM %s WHERE uid = ?", tableTexts),
		},
	},
	models.RecordBin: {
		actionCreate: {
			query: queryWithTable("INSERT INTO %s(uid, data, meta) VALUES (?, ?, ?)", tableBins),
		},
		actionRead: {
			query: queryWithTable("SELECT id, data, meta FROM %s WHERE id = ? AND uid = ?", tableBins),
		},
		actionUpdate: {
			query: queryWithTable("UPDATE %s SET data = ?, meta = ? WHERE id = ? AND uid = ?", tableBins),
		},
		actionDelete: {
			query: queryWithTable("DELETE FROM %s WHERE id = ? AND uid = ?", tableBins),
		},
		actionList: {
			query: queryWithTable("SELECT id, meta FROM %s WHERE uid = ?", tableBins),
		},
	},
	models.RecordBank: {
		actionCreate: {
			query: queryWithTable("INSERT INTO %s(uid, number, date, cvv, meta) VALUES (?, ?, ?, ?, ?)", tableBanks),
		},
		actionRead: {
			query: queryWithTable("SELECT id, number, date, cvv, meta FROM %s WHERE id = ? AND uid = ?", tableBanks),
		},
		actionUpdate: {
			query: queryWithTable("UPDATE %s SET number = ?, date =?, cvv = ? meta = ? WHERE id = ? AND uid = ?", tableBanks),
		},
		actionDelete: {
			query: queryWithTable("DELETE FROM %s WHERE id = ? AND uid = ?", tableBanks),
		},
		actionList: {
			query: queryWithTable("SELECT id, meta FROM %s WHERE uid = ?", tableBanks),
		},
	},
}
