package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sejo412/gophkeeper/internal/models"
)

type Storage struct {
	db *sql.DB
}

func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Init(ctx context.Context) error {
	queries := []query{
		{
			table: tableUsers,
			query: queryWithTable("CREATE TABLE IF NOT EXISTS %s(id INTEGER PRIMARY KEY, cn TEXT UNIQUE NOT NULL)",
				tableUsers),
		},
		{
			table: tablePasswords,
			query: queryWithTable("CREATE TABLE IF NOT EXISTS %s(id INTEGER PRIMARY KEY, login BLOB, password BLOB, meta BLOB)",
				tablePasswords),
		},
		{
			table: tableTexts,
			query: queryWithTable("CREATE TABLE IF NOT EXISTS %s(id INTEGER PRIMARY KEY, text BLOB, meta BLOB)",
				tableTexts),
		},
		{
			table: tableBins,
			query: queryWithTable("CREATE TABLE IF NOT EXISTS %s(id INTEGER PRIMARY KEY, data BLOB, meta BLOB)",
				tableBins),
		},
		{
			table: tableBanks,
			query: queryWithTable("CREATE TABLE IF NOT EXISTS %s(id INTEGER PRIMARY KEY, number BLOB, date BLOB, cvv BLOB, meta BLOB)",
				tableBanks),
		},
	}
	for _, t := range []table{tablePasswordsMap, tableTextsMap, tableBinsMap, tableBanksMap} {
		queries = append(queries, query{
			table: t,
			query: queryWithTable("CREATE TABLE IF NOT EXISTS %s(id INTEGER, uid INTEGER)", t),
		})
	}
	for _, q := range queries {
		if _, err := s.db.ExecContext(ctx, q.query); err != nil {
			return fmt.Errorf("failed create table %q: %w", q.table.String(), err)
		}
	}
	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) List(ctx context.Context, uid models.UserID) ([]models.RecordsEncrypted, error) {
	rec := make(map[models.RecordType]models.Encrypted)
	recs := make([]models.RecordsEncrypted, 0)
	for _, r := range []models.RecordType{
		models.RecordPassword,
		models.RecordText,
		models.RecordBin,
		models.RecordBank,
	} {
		m, err := s.getMap(ctx, uid, r)
		if err != nil {
			return nil, err
		}
		dataTable := tables(r).record
		q := query{
			table: dataTable,
			query: queryWithTable("SELECT COUNT(*) FROM %s WHERE uid=?", dataTable),
			args:  nil,
		}
	}
}

func (s *Storage) Get(ctx context.Context, user models.UserID, t models.RecordType,
	id models.ID) (*models.RecordEncrypted, error) {
	// TODO implement me
	panic("implement me")
}

func (s *Storage) Delete(ctx context.Context, user models.UserID, t models.RecordType, id models.ID) error {
	// TODO implement me
	panic("implement me")
}

func (s *Storage) Update(ctx context.Context, user models.UserID, t models.RecordType, id models.ID,
	record models.RecordEncrypted) error {
	// TODO implement me
	panic("implement me")
}

func (s *Storage) Add(ctx context.Context, userID models.UserID, t models.RecordType,
	record models.RecordEncrypted) error {
	if ok, err := s.IsUserExist(ctx, userID); err != nil || !ok {
		return fmt.Errorf("user id %q not exist or error: %w", userID, err)
	}
	switch t {
	case models.RecordPassword:
		return s.addPassword(ctx, userID, record.Password.Login, record.Password.Password, record.Password.Meta)
	default:
		return fmt.Errorf("invalid record type: %q", t)
	}

}

func (s *Storage) IsExist(ctx context.Context, uid models.UserID, t models.RecordType, id models.ID) (bool, error) {
	tableName := tables(t).record
	if tableName == tableUnknown {
		return false, fmt.Errorf("invalid record type: %q", t.String())
	}
	q := query{
		table: tableName,
		query: queryWithTable("SELECT COUNT(*) FROM %s WHERE id = ?", tableName),
		args:  []interface{}{id},
	}
	var count int
	if err := s.db.QueryRowContext(ctx, q.query, q.args...).Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		} else {
			return false, err
		}
	}
	ok, err := s.isOwner(ctx, uid, t, id)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *Storage) isOwner(ctx context.Context, uid models.UserID, t models.RecordType, id models.ID) (bool, error) {
	tableName := tables(t).recordMap
	if tableName == tableUnknown {
		return false, fmt.Errorf("invalid record type: %q", t.String())
	}
	q := query{
		table: tableName,
		query: queryWithTable("SELECT COUNT(*) FROM %s WHERE id = ? AND uid = ?", tableName),
		args:  []interface{}{id, uid},
	}
	var count int
	if err := s.db.QueryRowContext(ctx, q.query, q.args...).Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		} else {
			return false, err
		}
	}
	return count > 0, nil
}

func (s *Storage) Users(ctx context.Context) ([]*models.User, error) {
	q := query{
		table: tableUsers,
		query: queryWithTable("SELECT id, cn FROM %s ORDER BY id", tableUsers),
	}
	users := make([]*models.User, 0)
	rows, err := s.db.QueryContext(ctx, q.query)
	if err != nil {
		return users, fmt.Errorf("failed query users: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed get users: %w", err)
	}
	for rows.Next() {
		var id int
		var cn string
		if err = rows.Scan(&id, &cn); err != nil {
			return nil, fmt.Errorf("failed scan users: %w", err)
		}
		users = append(users, &models.User{
			ID: models.UserID(id),
			Cn: cn,
		})
	}

	return users, nil
}

func (s *Storage) NewUser(ctx context.Context, cn string) (*models.UserID, error) {
	q := query{
		table: tableUsers,
		query: queryWithTable("INSERT INTO %s(cn) VALUES (?) RETURNING id", tableUsers),
		args:  []interface{}{cn},
	}
	var uid int
	if err := s.db.QueryRowContext(ctx, q.query, q.args...).Scan(&uid); err != nil {
		return nil, fmt.Errorf("could not create user: %w", err)
	}
	result := models.UserID(uid)
	return &result, nil
}

func (s *Storage) IsUserExist(ctx context.Context, user models.UserID) (bool, error) {
	q := query{
		table: tableUsers,
		query: queryWithTable("SELECT COUNT(*) FROM %s WHERE id = ?", tableUsers),
		args:  []interface{}{user},
	}
	var count int
	if err := s.db.QueryRowContext(ctx, q.query, q.args...).Scan(&count); err != nil {
		return false, fmt.Errorf("failed check user exists: %w", err)
	}
	return count > 0, nil
}

func (s *Storage) userIDbyCn(ctx context.Context, cn string) (models.UserID, error) {
	q := query{
		table: tableUsers,
		query: queryWithTable("SELECT id FROM %s WHERE cn = ?", tableUsers),
		args:  []interface{}{cn},
	}
	var uid int
	err := s.db.QueryRowContext(ctx, q.query, q.args...).Scan(&uid)
	if errors.Is(err, sql.ErrNoRows) {
		return -1, nil
	}
	if err != nil {
		return -1, fmt.Errorf("could not get user by cn: %w", err)
	}
	result := models.UserID(uid)
	return result, nil
}

func (s *Storage) addPassword(ctx context.Context, uid models.UserID, login, password, meta []byte) error {
	q := query{
		table: tablePasswords,
		query: queryWithTable("INSERT INTO %s(login, password, meta) VALUES (?, ?, ?) RETURNING id", tablePasswords),
		args:  []interface{}{login, password, meta},
	}
	var id int
	if err := s.db.QueryRowContext(ctx, q.query, q.args...).Scan(&id); err != nil {
		return fmt.Errorf("failed add password for userID %d: %w", uid, err)
	}
	if err := s.addMap(ctx, models.ID(id), uid, tablePasswordsMap); err != nil {
		return fmt.Errorf("failed add mapping password for userID %d: %w", uid, err)
	}
	return nil
}

func (s *Storage) delPassword(ctx context.Context, uid models.UserID, password models.Password) error {
	if err := s.delMap(ctx, password.ID, uid, tablePasswordsMap); err != nil {
		return err
	}
	q := query{
		table: tablePasswords,
		query: queryWithTable("DELETE FROM %s WHERE id = ?", tablePasswords),
		args:  []interface{}{password.ID},
	}
	if _, err := s.db.ExecContext(ctx, q.query, q.args...); err != nil {
		return fmt.Errorf("failed delete from %q: %w", q.table.String(), err)
	}
	return nil
}

func (s *Storage) getPassword(ctx context.Context, uid models.UserID) ([]models.PasswordEncrypted, error) {
	m, err := s.getMap(ctx, uid, models.RecordPassword)
	if err != nil {
		return nil, err
	}
	q := query{
		table: tablePasswords,
	}
}

func (s *Storage) addMap(ctx context.Context, id models.ID, uid models.UserID, t table) error {
	q := query{
		table: t,
		query: queryWithTable("INSERT INTO %s(id, uid) VALUES(?, ?)", t),
		args:  []interface{}{id, uid},
	}
	_, err := s.db.ExecContext(ctx, q.query, q.args...)
	if err != nil {
		return fmt.Errorf("failed add map: %w", err)
	}
	return nil
}

func (s *Storage) delMap(ctx context.Context, id models.ID, uid models.UserID, t table) error {
	q := query{
		table: t,
		query: queryWithTable("DELETE FROM %s WHERE id = ? AND uid = ?", t),
		args:  []interface{}{id, uid},
	}
	_, err := s.db.ExecContext(ctx, q.query, q.args...)
	if err != nil {
		return fmt.Errorf("failed del map: %w", err)
	}
	return nil
}

func (s *Storage) getMap(ctx context.Context, uid models.UserID, r models.RecordType) ([]models.ID, error) {
	t := tables(r).recordMap
	if t == tableUnknown {
		return nil, fmt.Errorf("unknown record type: %v", r)
	}
	q := query{
		table: t,
		query: queryWithTable("SELECT id FROM %s WHERE uid = ?", t),
		args:  []interface{}{uid},
	}
	ids := make([]models.ID, 0)
	rows, err := s.db.QueryContext(ctx, q.query, q.args...)
	if err != nil {
		return nil, fmt.Errorf("failed query map: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed get map: %w", err)
	}
	for rows.Next() {
		var id models.ID
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed scan id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func queryWithTable(q string, t table) string {
	return fmt.Sprintf(q, t.String())
}
