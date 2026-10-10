package postgres

import (
	"context"

	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/store/postgres/sqlcgen"
)

// UserRepo implements ports.UserRepository.
type UserRepo struct{ db *DB }

var _ ports.UserRepository = (*UserRepo)(nil)

// NewUserRepo creates the repository.
func NewUserRepo(db *DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, u models.User) error {
	return mapErr(r.db.q.CreateUser(ctx, sqlcgen.CreateUserParams{
		ID:              string(u.ID),
		Email:           nullString(u.Email),
		Provider:        u.Provider,
		ProviderSubject: u.ProviderSubject,
		Nickname:        u.Nickname,
		AvatarUrl:       u.AvatarURL,
	}))
}

func (r *UserRepo) ByID(ctx context.Context, id models.UserID) (models.User, error) {
	row, err := r.db.q.GetUser(ctx, string(id))
	if err != nil {
		return models.User{}, mapErr(err)
	}
	return toUser(row), nil
}

func (r *UserRepo) ByEmail(ctx context.Context, email string) (models.User, error) {
	row, err := r.db.q.GetUserByEmail(ctx, nullString(&email))
	if err != nil {
		return models.User{}, mapErr(err)
	}
	return toUser(row), nil
}

func (r *UserRepo) ByIdentity(ctx context.Context, provider, subject string) (models.User, error) {
	row, err := r.db.q.GetUserByIdentity(ctx, sqlcgen.GetUserByIdentityParams{Provider: provider, ProviderSubject: subject})
	if err != nil {
		return models.User{}, mapErr(err)
	}
	return toUser(row), nil
}

func (r *UserRepo) Delete(ctx context.Context, id models.UserID) error {
	return mapErr(r.db.q.DeleteUser(ctx, string(id)))
}

func toUser(row sqlcgen.User) models.User {
	u := models.User{
		ID:              models.UserID(row.ID),
		Provider:        row.Provider,
		ProviderSubject: row.ProviderSubject,
		Nickname:        row.Nickname,
		AvatarURL:       row.AvatarUrl,
		CreatedAt:       row.CreatedAt,
	}
	if row.Email.Valid {
		email := row.Email.String
		u.Email = &email
	}
	return u
}

func (r *UserRepo) UpdateAvatar(ctx context.Context, id models.UserID, url string) error {
	return mapErr(r.db.q.UpdateUserAvatar(ctx, sqlcgen.UpdateUserAvatarParams{ID: string(id), AvatarUrl: url}))
}
