package users

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"

	"github.com/Masterminds/squirrel"
	"github.com/bnkamalesh/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgstore struct {
	qbuilder  squirrel.StatementBuilderType
	pqdriver  *pgxpool.Pool
	tableName string
}

func (ps *pgstore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query, args, err := ps.qbuilder.Select(
		"id",
		"name",
		"email",
		"phone",
		"address",
	).From(
		ps.tableName,
	).Where(
		squirrel.Eq{"email": email},
	).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed preparing query")
	}

	user := new(User)
	address := new(sql.NullString)
	phone := new(sql.NullString)
	row := ps.pqdriver.QueryRow(ctx, query, args...)
	err = row.Scan(user.ID, user.Name, phone, address)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting user info")
	}
	user.Address = address.String
	user.Phone = phone.String

	return user, nil
}

func (ps *pgstore) SaveUser(ctx context.Context, user *User) (string, error) {
	user.ID = ps.newUserID(user.Name, user.Email)

	query, args, err := ps.qbuilder.Insert(
		ps.tableName,
	).Columns(
		"id",
		"name",
		"email",
		"phone",
		"address",
	).Values(
		user.ID,
		user.Name,
		user.Email,
		sql.NullString{
			String: user.Phone,
			Valid:  len(user.Phone) == 0,
		},
		sql.NullString{
			String: user.Address,
			Valid:  len(user.Address) == 0,
		},
	).ToSql()
	if err != nil {
		return "", errors.Wrap(err, "failed preparing query")
	}
	_, err = ps.pqdriver.Exec(ctx, query, args...)
	if err != nil {
		return "", errors.Wrap(err, "failed storing user info")
	}

	return user.ID, nil
}

func (ps *pgstore) BulkSaveUser(ctx context.Context, users []User) error {
	rows := make([][]any, 0, len(users))

	for _, user := range users {
		rows = append(rows, []any{
			user.ID,
			user.Name,
			user.Email,
			sql.NullString{
				String: user.Phone,
				Valid:  len(user.Phone) == 0,
			},
			sql.NullString{
				String: user.Address,
				Valid:  len(user.Address) == 0,
			},
		})
	}

	inserted, err := ps.pqdriver.CopyFrom(
		ctx,
		pgx.Identifier{ps.tableName},
		[]string{"id", "name", "email", "phone", "address"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return errors.Wrap(err, "failed inserting users")
	}

	ulen := int64(len(users))
	if inserted != ulen {
		return errors.Internalf(
			"failed inserting %d out of %d users",
			ulen-inserted,
			ulen,
		)
	}

	return nil
}

func (ps *pgstore) newUserID(name, email string) string {
	b := bytes.NewBufferString(name + email)
	chksum := sha256.Sum224(b.Bytes())
	return hex.EncodeToString(chksum[:])
}

func NewPostgresStore(pqdriver *pgxpool.Pool, tablename string) *pgstore {
	return &pgstore{
		qbuilder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		pqdriver:  pqdriver,
		tableName: tablename,
	}
}
