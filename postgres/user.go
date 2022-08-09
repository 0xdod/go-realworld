package postgres

import (
	"context"
	"fmt"
	"log"

	"github.com/0xdod/go-realworld/conduit"
	"github.com/jmoiron/sqlx"
)

type UserService struct {
	db *DB
}

func NewUserService(db *DB) *UserService {
	return &UserService{db}
}

func (us *UserService) CreateUser(ctx context.Context, user *conduit.User) error {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := createUser(ctx, tx, user); err != nil {
		return err
	}

	return tx.Commit()
}

func (us *UserService) UserByID(ctx context.Context, id uint) (*conduit.User, error) {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	user, err := findOneUser(ctx, tx, conduit.UserFilter{ID: &id})

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func (us *UserService) UserByEmail(ctx context.Context, email string) (*conduit.User, error) {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	user, err := findOneUser(ctx, tx, conduit.UserFilter{Email: &email})

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func (us *UserService) UserByUsername(ctx context.Context, uname string) (*conduit.User, error) {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	user, err := findOneUser(ctx, tx, conduit.UserFilter{Username: &uname})

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func (us *UserService) Users(ctx context.Context, uf conduit.UserFilter) ([]*conduit.User, error) {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	users, err := findUsers(ctx, tx, uf)

	if err != nil {
		return nil, err
	}

	return users, tx.Commit()
}

func (us *UserService) Authenticate(ctx context.Context, email, password string) (*conduit.User, error) {
	user, err := us.UserByEmail(ctx, email)

	if err != nil {
		return nil, err
	}

	if !user.VerifyPassword(password) {
		return nil, conduit.ErrUnAuthorized
	}

	return user, nil
}

func (us *UserService) UpdateUser(ctx context.Context, user *conduit.User, patch conduit.UserPatch) error {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		log.Println(err)
		return conduit.ErrInternal
	}

	defer tx.Rollback()

	if err := updateUser(ctx, tx, user, patch); err != nil {
		log.Println(err)
		return conduit.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return conduit.ErrInternal
	}

	return nil
}

func (us *UserService) DeleteUser(ctx context.Context, id uint) error {
	return nil
}

// FollowUser follower -> follows -> user
func (us *UserService) FollowUser(ctx context.Context, user, follower *conduit.User) error {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err = addFollowing(ctx, tx, user, follower); err != nil {
		return err
	}

	return tx.Commit()
}

func (us *UserService) UnFollowUser(ctx context.Context, user, follower *conduit.User) error {
	tx, err := us.db.BeginTxx(ctx, nil)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err = removeFollowing(ctx, tx, user, follower); err != nil {
		return err
	}

	return tx.Commit()
}

func createUser(ctx context.Context, tx *sqlx.Tx, user *conduit.User) error {
	query := `
	INSERT INTO users (email, username, bio, image, password_hash)
	VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at
	`
	args := []interface{}{user.Email, user.Username, user.Bio, user.Image, user.PasswordHash}
	err := tx.QueryRowxContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return conduit.ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return conduit.ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func findUserByID(ctx context.Context, tx *sqlx.Tx, id uint) (*conduit.User, error) {
	return findOneUser(ctx, tx, conduit.UserFilter{ID: &id})
}

func findOneUser(ctx context.Context, tx *sqlx.Tx, filter conduit.UserFilter) (*conduit.User, error) {
	us, err := findUsers(ctx, tx, filter)

	if err != nil {
		return nil, err
	} else if len(us) == 0 {
		return nil, conduit.ErrNotFound
	}

	return us[0], nil
}

func findUsers(ctx context.Context, tx *sqlx.Tx, filter conduit.UserFilter) ([]*conduit.User, error) {
	where, args := []string{}, []interface{}{}
	argPosition := 0

	if v := filter.ID; v != nil {
		argPosition++
		where, args = append(where, fmt.Sprintf("id = $%d", argPosition)), append(args, *v)
	}

	if v := filter.Email; v != nil {
		argPosition++
		where, args = append(where, fmt.Sprintf("email = $%d", argPosition)), append(args, *v)
	}

	if v := filter.Username; v != nil {
		argPosition++
		where, args = append(where, fmt.Sprintf("username = $%d", argPosition)), append(args, *v)
	}

	query := "SELECT * from users" + formatWhereClause(where) +
		" ORDER BY id ASC" + formatLimitOffset(filter.Limit, filter.Offset)

	users, err := queryUsers(ctx, tx, query, args...)

	if err != nil {
		return nil, err
	}

	for _, user := range users {
		followers, _ := getFollowers(ctx, tx, user)
		user.Followers = followers
	}

	return users, nil
}

func updateUser(ctx context.Context, tx *sqlx.Tx, user *conduit.User, patch conduit.UserPatch) error {
	if v := patch.Bio; v != nil {
		user.Bio = *v
	}

	if v := patch.Email; v != nil {
		user.Email = *v
	}

	if v := patch.PasswordHash; v != nil {
		user.PasswordHash = *v
	}

	if v := patch.Image; v != nil {
		user.Image = *v
	}

	if v := patch.Username; v != nil {
		user.Username = *v
	}

	args := []interface{}{
		user.Username,
		user.Email,
		user.Bio,
		user.Image,
		user.PasswordHash,
		user.ID,
	}

	query := `
	UPDATE users 
	SET username = $1, email = $2, bio = $3, image = $4, password_hash = $5, updated_at = NOW()
	WHERE id = $6
	RETURNING updated_at`

	if err := tx.QueryRowxContext(ctx, query, args...).Scan(&user.UpdatedAt); err != nil {
		log.Printf("error updating record: %v", err)
		return conduit.ErrInternal
	}

	return nil
}

func addFollowing(ctx context.Context, tx *sqlx.Tx, following, follower *conduit.User) error {
	query := "INSERT INTO followings (following_id, follower_id) VALUES ($1, $2)"
	return execQuery(ctx, tx, query, following.ID, follower.ID)
}

func removeFollowing(ctx context.Context, tx *sqlx.Tx, following, follower *conduit.User) error {
	query := "DELETE FROM followings WHERE following_id = $1 AND follower_id = $2"
	return execQuery(ctx, tx, query, following.ID, follower.ID)
}

func getFollowers(ctx context.Context, tx *sqlx.Tx, user *conduit.User) ([]*conduit.User, error) {
	query := `SELECT * FROM users WHERE id IN (
		SELECT follower_id FROM followings WHERE following_id = $1
	)
	`
	return queryUsers(ctx, tx, query, user.ID)
}

func queryUsers(ctx context.Context, tx *sqlx.Tx, query string, args ...interface{}) ([]*conduit.User, error) {
	users := make([]*conduit.User, 0)

	if err := findMany(ctx, tx, &users, query, args...); err != nil {
		return users, err
	}

	return users, nil
}
