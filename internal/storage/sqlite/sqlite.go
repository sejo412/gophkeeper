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
			query: queryWithTable("CREATE TABLE IF NOT EXISTS %s(id INTEGER PRIMARY KEY, uid INTEGER NOT NULL, login BLOB, password BLOB, meta BLOB)",
				tablePasswords),
		},
		{
			table: tableTexts,
			query: queryWithTable("CREATE TABLE IF NOT EXISTS %s(id INTEGER PRIMARY KEY, uid INTEGER NOT NULL, text BLOB, meta BLOB)",
				tableTexts),
		},
		{
			table: tableBins,
			query: queryWithTable("CREATE TABLE IF NOT EXISTS %s(id INTEGER PRIMARY KEY, uid INTEGER NOT NULL, data BLOB, meta BLOB)",
				tableBins),
		},
		{
			table: tableBanks,
			query: queryWithTable("CREATE TABLE IF NOT EXISTS %s(id INTEGER PRIMARY KEY, uid, INTEGER NOT NULL, number BLOB, date BLOB, cvv BLOB, meta BLOB)",
				tableBanks),
		},
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

func (s *Storage) List(ctx context.Context, uid models.UserID) (models.RecordsEncrypted, error) {
	var recs models.RecordsEncrypted
	passwords, err := s.getPasswords(ctx, uid)
	if err != nil {
		return recs, err
	}
	recs.Password = passwords
	return recs, nil
}

func (s *Storage) Get(ctx context.Context, user models.UserID, t models.RecordType,
	id models.ID) (models.RecordEncrypted, error) {
	var rec models.RecordEncrypted
	switch t {
	case models.RecordPassword:
		passwords, err := s.getPasswords(ctx, user)
		if err != nil || len(passwords) != 1 {
			return rec, err
		}
		rec.Password = passwords[0]
		return rec, nil
	default:
		return rec, errors.New("invalid record type")
	}
}

func (s *Storage) Delete(ctx context.Context, uid models.UserID, t models.RecordType, id models.ID) error {
	switch t {
	case models.RecordPassword:
		return s.delPassword(ctx, uid, models.Password{ID: id})
	default:
		return errors.New("invalid record type")
	}
}

func (s *Storage) Update(ctx context.Context, uid models.UserID, t models.RecordType, id models.ID,
	record models.RecordEncrypted) error {
	switch t {
	case models.RecordPassword:
		return s.updatePassword(ctx, uid, id, record.Password.Login, record.Password.Password, record.Password.Meta)
	default:
		return errors.New("invalid record type")
	}
}

func (s *Storage) Add(ctx context.Context, uid models.UserID, t models.RecordType,
	record models.RecordEncrypted) error {
	if ok, err := s.IsUserExist(ctx, uid); err != nil || !ok {
		return fmt.Errorf("user id %q not exist or error: %w", uid, err)
	}
	switch t {
	case models.RecordPassword:
		return s.addPassword(ctx, uid, record.Password.Login, record.Password.Password, record.Password.Meta)
	default:
		return fmt.Errorf("invalid record type: %q", t)
	}

}

func (s *Storage) IsExist(ctx context.Context, uid models.UserID, t models.RecordType, id models.ID) (bool, error) {
	tableName := tables(t)
	if tableName == tableUnknown {
		return false, fmt.Errorf("invalid record type: %q", t.String())
	}
	q := query{
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
	tableName := tables(t)
	if tableName == tableUnknown {
		return false, fmt.Errorf("invalid record type: %q", t.String())
	}
	q := query{
		query: queryWithTable("SELECT uid FROM %s WHERE id = ?", tableName),
		args:  []interface{}{id, uid},
	}
	var realUID int
	if err := s.db.QueryRowContext(ctx, q.query, q.args...).Scan(&realUID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		} else {
			return false, err
		}
	}
	return realUID == int(uid), nil
}

func (s *Storage) Users(ctx context.Context) ([]*models.User, error) {
	q := query{
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
		query: queryWithTable("INSERT INTO %s(uid, login, password, meta) VALUES (?, ?, ?, ?)", tablePasswords),
		args:  []interface{}{uid, login, password, meta},
	}
	if _, err := s.db.ExecContext(ctx, q.query, q.args...); err != nil {
		return fmt.Errorf("failed add password for userID %d: %w", uid, err)
	}
	return nil
}

func (s *Storage) delPassword(ctx context.Context, uid models.UserID, password models.Password) error {
	q := query{
		query: queryWithTable("DELETE FROM %s WHERE id = ? AND uid = ?", tablePasswords),
		args:  []interface{}{password.ID, uid},
	}
	res, err := s.db.ExecContext(ctx, q.query, q.args...)
	if err != nil {
		return fmt.Errorf("failed delete from %q: %w", q.table.String(), err)
	}
	rowsCount, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed get rows affected: %w", err)
	}
	if rowsCount == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func (s *Storage) getPasswords(ctx context.Context, uid models.UserID) ([]models.PasswordEncrypted, error) {
	q := query{
		query: queryWithTable("SELECT id, login, password, meta FROM %s WHERE uid = ?", tablePasswords),
		args:  []interface{}{uid},
	}
	results := make([]models.PasswordEncrypted, 0)
	rows, err := s.db.QueryContext(ctx, q.query, q.args...)
	if err != nil {
		return results, fmt.Errorf("failed query passwords: %w", err)
	}
	if rows.Err() != nil {
		return results, fmt.Errorf("failed get passwords: %w", rows.Err())
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		rec := &models.PasswordEncrypted{}
		if err := rows.Scan(&rec.ID, &rec.Login, &rec.Password, &rec.Meta); err != nil {
			return results, fmt.Errorf("failed scan passwords: %w", err)
		}
		results = append(results, *rec)
	}
	return results, nil
}

func (s *Storage) updatePassword(ctx context.Context, uid models.UserID, id models.ID,
	newLogin, newPassword, newMeta []byte) error {
	q := query{
		query: queryWithTable("UPDATE %s SET login = ?, password = ?, meta = ? WHERE uid = ? AND id = ?",
			tablePasswords),
		args: []interface{}{newLogin, newPassword, newMeta, uid, id},
	}
	res, err := s.db.ExecContext(ctx, q.query, q.args...)
	if err != nil {
		return fmt.Errorf("failed update password for userID %d: %w", uid, err)
	}
	rowCount, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed get rows affected: %w", err)
	}
	if rowCount == 0 {
		return fmt.Errorf("no rows affected for userID %d", uid)
	}
	return nil
}

func queryWithTable(q string, t table) string {
	return fmt.Sprintf(q, t.String())
}
