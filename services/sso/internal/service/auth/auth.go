package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/lib/logger/sl"
	"sso/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	userDeleter  UserDeleter
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveStudent(ctx context.Context,
		email string,
		passHash []byte,
		name string,
		surname string,
		middleName string,
		phoneNumber string,
		birthDate models.BirthDate,
		group string,
		studentNumber string,
	) (userID int64, err error)
	SaveTeacher(ctx context.Context,
		email string,
		passHash []byte,
		name string,
		surname string,
		middleName string,
		phoneNumber string,
		birthDate models.BirthDate,
		title string,
		department string,
		degree string,
	) (userID int64, err error)
	SaveAdmin(ctx context.Context,
		email string,
		passHash []byte,
	) (userID int64, err error)
}
type UserDeleter interface {
	DeleteStudent(ctx context.Context, userID int64) error
	DeleteTeacher(ctx context.Context, userID int64) error
	DeleteAdmin(ctx context.Context, userID int64) error
}
type UserProvider interface {
	Student(ctx context.Context, email string) (models.Student, error)
	IsStudent(ctx context.Context, userID int64) (bool, error)
	Teacher(ctx context.Context, email string) (models.Teacher, error)
	IsTeacher(ctx context.Context, userID int64) (bool, error)
	Admin(ctx context.Context, email string) (models.Admin, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidUserID      = errors.New("invalid userID")
	ErrUserExists         = errors.New("user already exists")
	ErrRoleNotExists      = errors.New("role does not exists")
)

const emptyID = 0

// New returns a new instance of the Auth service.
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	userDeleter UserDeleter,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		userDeleter:  userDeleter,
		tokenTTL:     tokenTTL,
	}
}

// Login checks if user exists and returns token
func (a *Auth) Login(ctx context.Context,
	email string,
	password string,
	role string,
) (string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("user_email", email),
	)

	log.Info("start login user")

	var (
		user interface{}
		err  error
	)

	switch role {
	case "student":
		user, err = a.userProvider.Student(ctx, email)
	case "teacher":
		user, err = a.userProvider.Teacher(ctx, email)
	case "admin":
		user, err = a.userProvider.Admin(ctx, email)
	default:
		return "", fmt.Errorf("%s: %w", op, ErrRoleNotExists)
	}
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err = bcrypt.CompareHashAndPassword(user.(models.User).PassHash, []byte(password)); err != nil {
		a.log.Warn("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	token, err := jwt.NewToken(user.(models.User), a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user is successfully logged in")

	return token, nil
}

func (a *Auth) RegisterTeacher(
	ctx context.Context,
	email string,
	password string,
	name string,
	surname string,
	middleName string,
	phoneNumber string,
	birthDate models.BirthDate,
	title string,
	department string,
	degree string,
) (userID int64, err error) {
	const op = "auth.RegisterTeacher"

	log := a.log.With(
		slog.String("op", op),
		slog.String("user_email", email),
	)

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate hash", sl.Err(err))

		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.userSaver.SaveTeacher(ctx, email, passHash, name, surname, middleName, phoneNumber, birthDate, title, department, degree)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("teacher already exists", sl.Err(err))

			return emptyID, fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed to save teacher", sl.Err(err))

		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("teacher registered")

	return id, nil
}

func (a *Auth) RegisterStudent(
	ctx context.Context,
	email string,
	password string,
	name string,
	surname string,
	middleName string,
	phoneNumber string,
	birthDate models.BirthDate,
	group string,
	studentNumber string,
) (userID int64, err error) {
	const op = "auth.RegisterStudent"

	log := a.log.With(
		slog.String("op", op),
		slog.String("user_email", email),
	)

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate hash", sl.Err(err))

		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.userSaver.SaveStudent(ctx, email, passHash, name, surname, middleName, phoneNumber, birthDate, group, studentNumber)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("student already exists", sl.Err(err))

			return emptyID, fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed to save student", sl.Err(err))

		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("student registered")

	return id, nil
}

func (a *Auth) RegisterAdmin(
	ctx context.Context,
	email string,
	password string,
) (userID int64, err error) {
	const op = "auth.RegisterAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.String("user_email", email),
	)

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate hash", sl.Err(err))

		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.userSaver.SaveAdmin(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("admin already exists", sl.Err(err))

			return emptyID, fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed to save admin", sl.Err(err))

		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("student registered")

	return id, nil
}

func (a *Auth) DeleteStudent(ctx context.Context, userID int64) error {
	const op = "auth.DeleteStudent"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("user_id", int(userID)),
	)

	err := a.userDeleter.DeleteStudent(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("student not found", sl.Err(err))

			return fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed to delete student", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("student deleted")

	return nil
}

func (a *Auth) DeleteTeacher(ctx context.Context, userID int64) error {
	const op = "auth.DeleteTeacher"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("user_id", int(userID)),
	)

	err := a.userDeleter.DeleteTeacher(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("teacher not found", sl.Err(err))

			return fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed to delete Teacher", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("teacher deleted")

	return nil
}

func (a *Auth) DeleteAdmin(ctx context.Context, userID int64) error {
	const op = "auth.DeleteAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("user_id", int(userID)),
	)

	err := a.userDeleter.DeleteStudent(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("admin not found", sl.Err(err))

			return fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed to admin", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("admin deleted")

	return nil
}

// IsAdmin checks if user is admin
func (a *Auth) IsAdmin(ctx context.Context,
	userID int64,
) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("user_id", int(userID)),
	)

	isAdmin, err := a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
		}

		log.Error("failed to check is user an admin")

		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil

}

func (a *Auth) IsStudent(ctx context.Context,
	userID int64,
) (bool, error) {
	const op = "auth.IsStudent"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("user_id", int(userID)),
	)

	isStudent, err := a.userProvider.IsStudent(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
		}

		log.Error("failed to check is user a Student")

		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully checked if user is student", slog.Bool("is_student", isStudent))

	return isStudent, nil

}

func (a *Auth) IsTeacher(ctx context.Context,
	userID int64,
) (bool, error) {
	const op = "auth.IsTeacher"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("user_id", int(userID)),
	)

	isTeacher, err := a.userProvider.IsTeacher(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
		}

		log.Error("failed to check is user a Teacher")

		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully checked if user is teacher", slog.Bool("is_teacher", isTeacher))

	return isTeacher, nil

}
