package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sejo412/gophkeeper/internal/models"
)

// Storage implements server.Storage interface.
type Storage struct {
	db *sql.DB
}

// New constructs Storage object.
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

// Init creates tables in database.
func (s *Storage) Init(ctx context.Context) error {
	queries := []query{
		{
			table: tableUsers,
			query: queryWithTable(
				"CREATE TABLE IF NOT EXISTS %s(id INTEGER PRIMARY KEY, cn TEXT UNIQUE NOT NULL)",
				tableUsers,
			),
		},
		{
			table: tablePasswords,
			query: queryWithTable(
				"CREATE TABLE IF NOT EXISTS %s(id INTEGER PRIMARY KEY, uid INTEGER NOT NULL, login BLOB, password BLOB, meta BLOB)",
				tablePasswords,
			),
		},
		{
			table: tableTexts,
			query: queryWithTable(
				"CREATE TABLE IF NOT EXISTS %s(id INTEGER PRIMARY KEY, uid INTEGER NOT NULL, text BLOB, meta BLOB)",
				tableTexts,
			),
		},
		{
			table: tableBins,
			query: queryWithTable(
				"CREATE TABLE IF NOT EXISTS %s(id INTEGER PRIMARY KEY, uid INTEGER NOT NULL, data BLOB, meta BLOB)",
				tableBins,
			),
		},
		{
			table: tableBanks,
			query: queryWithTable(
				"CREATE TABLE IF NOT EXISTS %s(id INTEGER PRIMARY KEY, uid INTEGER NOT NULL, number BLOB, "+
					"name BLOB, date BLOB, cvv BLOB, meta BLOB)",
				tableBanks,
			),
		},
	}
	for _, q := range queries {
		if _, err := s.db.ExecContext(ctx, q.query); err != nil {
			return fmt.Errorf("failed create table %q: %w", q.table.String(), err)
		}
	}
	return nil
}

// Close closes storage.
func (s *Storage) Close() error {
	return s.db.Close()
}

// ListAll returns id, meta for all records.
func (s *Storage) ListAll(ctx context.Context, uid models.UserID) (models.RecordsEncrypted, error) {
	result := models.RecordsEncrypted{}
	for _, recType := range []models.RecordType{
		models.RecordPassword,
		models.RecordText,
		models.RecordBin,
		models.RecordBank,
	} {
		records, err := s.List(ctx, uid, recType)
		if err != nil {
			return models.RecordsEncrypted{}, fmt.Errorf("failed to list %q: %w", recType.String(), err)
		}
		switch recType {
		case models.RecordPassword:
			result.Password = records.Password
		case models.RecordText:
			result.Text = records.Text
		case models.RecordBin:
			result.Bin = records.Bin
		case models.RecordBank:
			result.Bank = records.Bank
		default:
		}
	}
	return result, nil
}

// List returns records id, meta by type.
func (s *Storage) List(ctx context.Context, uid models.UserID, t models.RecordType) (models.RecordsEncrypted, error) {
	args := []interface{}{uid}
	rows, err := s.db.QueryContext(ctx, actions[t][actionList].query, args...)
	if err != nil {
		return models.RecordsEncrypted{}, fmt.Errorf("failed query %q: %w", t.String(), err)
	}
	if rows.Err() != nil {
		return models.RecordsEncrypted{}, fmt.Errorf("failed get %s: %w", t.String(), rows.Err())
	}
	defer func() {
		_ = rows.Close()
	}()
	result := models.RecordsEncrypted{
		Password: []models.PasswordEncrypted{},
		Text:     []models.TextEncrypted{},
		Bin:      []models.BinEncrypted{},
		Bank:     []models.BankEncrypted{},
	}
	for rows.Next() {
		var id models.ID
		var meta []byte
		if err = rows.Scan(&id, &meta); err != nil {
			return models.RecordsEncrypted{}, fmt.Errorf("failed scan %q: %w", t.String(), err)
		}
		switch t {
		case models.RecordPassword:
			result.Password = append(
				result.Password, models.PasswordEncrypted{
					ID:   id,
					Meta: meta,
				},
			)
		case models.RecordText:
			result.Text = append(
				result.Text, models.TextEncrypted{
					ID:   id,
					Meta: meta,
				},
			)
		case models.RecordBin:
			result.Bin = append(
				result.Bin, models.BinEncrypted{
					ID:   id,
					Meta: meta,
				},
			)
		case models.RecordBank:
			result.Bank = append(
				result.Bank, models.BankEncrypted{
					ID:   id,
					Meta: meta,
				},
			)
		default:
		}
	}
	if err = rows.Err(); err != nil {
		return models.RecordsEncrypted{}, fmt.Errorf("failed iterate %s: %w", t.String(), rows.Err())
	}
	return result, nil
}

// Get returns object by id, userid and record type.
func (s *Storage) Get(
	ctx context.Context, uid models.UserID, t models.RecordType,
	id models.ID,
) (models.RecordEncrypted, error) {
	args := []interface{}{id, uid}
	rec := models.RecordEncrypted{
		Password: models.PasswordEncrypted{},
		Text:     models.TextEncrypted{},
		Bin:      models.BinEncrypted{},
		Bank:     models.BankEncrypted{},
	}
	row := s.db.QueryRowContext(ctx, actions[t][actionRead].query, args...)
	var err error
	switch t {
	case models.RecordPassword:
		err = row.Scan(&rec.Password.ID, &rec.Password.Login, &rec.Password.Password, &rec.Password.Meta)
	case models.RecordText:
		err = row.Scan(&rec.Text.ID, &rec.Text.Text, &rec.Text.Meta)
	case models.RecordBin:
		err = row.Scan(&rec.Bin.ID, &rec.Bin.Data, &rec.Bin.Meta)
	case models.RecordBank:
		err = row.Scan(&rec.Bank.ID, &rec.Bank.Number, &rec.Bank.Name, &rec.Bank.Date, &rec.Bank.Cvv, &rec.Bank.Meta)
	default:
		err = errors.New("unknown record")
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.RecordEncrypted{}, fmt.Errorf("%q with %d not found", t.String(), id)
		} else {
			return models.RecordEncrypted{}, err
		}
	}
	return rec, nil
}

// Delete deletes object by id, userid and record type.
func (s *Storage) Delete(ctx context.Context, uid models.UserID, t models.RecordType, id models.ID) error {
	args := []interface{}{id, uid}
	res, err := s.db.ExecContext(ctx, actions[t][actionDelete].query, args...)
	if err != nil {
		return fmt.Errorf("failed delete %q with id %d: %w", t.String(), id, err)
	}
	rowsCount, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed get rows affected: %w", err)
	}
	if rowsCount == 0 {
		return fmt.Errorf("nothing to delete")
	}
	return nil
}

// Update updates object by id, userid and record type.
func (s *Storage) Update(
	ctx context.Context, uid models.UserID, t models.RecordType, id models.ID,
	record models.RecordEncrypted,
) error {
	var args []interface{}
	switch t {
	case models.RecordPassword:
		args = []interface{}{record.Password.Login, record.Password.Password, record.Password.Meta, id, uid}
	case models.RecordText:
		args = []interface{}{record.Text.Text, record.Text.Meta, id, uid}
	case models.RecordBin:
		args = []interface{}{record.Bin.Data, record.Bin.Meta, id, uid}
	case models.RecordBank:
		args = []interface{}{
			record.Bank.Number, record.Bank.Name, record.Bank.Date, record.Bank.Cvv, record.Bank.Meta,
			id, uid,
		}
	default:
		return errors.New("invalid record type")
	}
	res, err := s.db.ExecContext(ctx, actions[t][actionUpdate].query, args...)
	if err != nil {
		return fmt.Errorf("failed update %q for userID %d: %w", t.String(), uid, err)
	}
	rowCount, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed get rows affected: %w", err)
	}
	if rowCount == 0 {
		return fmt.Errorf("no records updated for userID %d", uid)
	}
	return nil
}

// Add adds encrypted record for user by record type.
func (s *Storage) Add(
	ctx context.Context, uid models.UserID, t models.RecordType,
	record models.RecordEncrypted,
) error {
	if ok, err := s.IsUserExist(ctx, uid); err != nil || !ok {
		return fmt.Errorf("user id %q not exist or error: %w", uid, err)
	}
	var args []interface{}
	switch t {
	case models.RecordPassword:
		args = []interface{}{uid, record.Password.Login, record.Password.Password, record.Password.Meta}
	case models.RecordText:
		args = []interface{}{uid, record.Text.Text, record.Text.Meta}
	case models.RecordBin:
		args = []interface{}{uid, record.Bin.Data, record.Bin.Meta}
	case models.RecordBank:
		args = []interface{}{uid, record.Bank.Number, record.Bank.Name, record.Bank.Date, record.Bank.Cvv, record.Bank.Meta}
	default:
		return fmt.Errorf("invalid record type: %q", t)
	}
	if _, err := s.db.ExecContext(ctx, actions[t][actionCreate].query, args...); err != nil {
		return fmt.Errorf("failed create record %q for %d: %w", t.String(), uid, err)
	}
	return nil
}

// IsExist returns true if record exists.
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

// Users returns all registered users.
func (s *Storage) Users(ctx context.Context) ([]models.User, error) {
	q := query{
		query: queryWithTable("SELECT id, cn FROM %s ORDER BY id", tableUsers),
	}
	users := make([]models.User, 0)
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
		users = append(
			users, models.User{
				ID: models.UserID(id),
				Cn: cn,
			},
		)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterate users: %w", err)
	}

	return users, nil
}

// NewUser creates new user.
func (s *Storage) NewUser(ctx context.Context, cn string) (models.UserID, error) {
	q := query{
		query: queryWithTable("INSERT INTO %s(cn) VALUES (?) RETURNING id", tableUsers),
		args:  []interface{}{cn},
	}
	var uid int
	if err := s.db.QueryRowContext(ctx, q.query, q.args...).Scan(&uid); err != nil {
		return -1, fmt.Errorf("could not create user: %w", err)
	}
	result := models.UserID(uid)
	return result, nil
}

// IsUserExist returns true if user exists in database.
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

// GetUserID returns uid by common name.
func (s *Storage) GetUserID(ctx context.Context, cn string) (models.UserID, error) {
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

func queryWithTable(q string, t table) string {
	return fmt.Sprintf(q, t.String())
}
