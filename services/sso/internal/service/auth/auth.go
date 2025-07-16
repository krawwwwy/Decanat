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
	log             *slog.Logger
	pendingSaver    PendingSaver
	pendingProvider PendingProvider
	pendingDeleter  PendingDeleter
	approver        Approver
	userProvider    UserProvider
	userDeleter     UserDeleter
	tokenTTL        time.Duration
}

type PendingSaver interface {
	SavePendingUser(ctx context.Context,
		email string,
		passHash []byte,
		name string,
		surname string,
		middleName string,
		phoneNumber string,
		birthDate models.BirthDate,
		meta map[string]string,
	) error
}
type Approver interface {
	ApprovePendingUser(ctx context.Context, userID string) error
}

type PendingProvider interface {
	ListPendingUsers(ctx context.Context) ([]models.PendingUser, error)
}

type PendingDeleter interface {
	DeletePendingUser(ctx context.Context, id string) error
}

type UserDeleter interface {
	DeleteUser(ctx context.Context, userID int64, role string) (err error)
}
type UserProvider interface {
	Student(ctx context.Context, email string) (models.Student, error)
	Teacher(ctx context.Context, email string) (models.Teacher, error)
	Admin(ctx context.Context, email string) (models.Admin, error)

	StudentsList(ctx context.Context) ([]models.Student, error)
	TeacherList(ctx context.Context) ([]models.Teacher, error)
	AdminList(ctx context.Context) ([]models.Admin, error)

	IsStudent(ctx context.Context, userID int64) (bool, error)
	IsTeacher(ctx context.Context, userID int64) (bool, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidUserID       = errors.New("invalid userID")
	ErrUserExists          = errors.New("user already exists")
	ErrRoleNotExists       = errors.New("role does not exists")
	ErrUserNotMatchingRole = errors.New("user not matching role")
)

const emptyID = 0

// New returns a new instance of the Auth service.
func New(
	log *slog.Logger,
	pendingSaver PendingSaver,
	pendingProvider PendingProvider,
	pendingDeleter PendingDeleter,
	approver Approver,
	userProvider UserProvider,
	userDeleter UserDeleter,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:             log,
		pendingSaver:    pendingSaver,
		pendingProvider: pendingProvider,
		pendingDeleter:  pendingDeleter,
		approver:        approver,
		userProvider:    userProvider,
		userDeleter:     userDeleter,
		tokenTTL:        tokenTTL,
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
		if errors.Is(err, storage.ErrUserHasAnotherRole) {
			a.log.Warn("try to log with another role", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrUserNotMatchingRole)
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

func (a *Auth) RegisterPending(
	ctx context.Context,
	email string,
	password string,
	name string,
	surname string,
	middleName string,
	phoneNumber string,
	birthDate models.BirthDate,
	meta map[string]string,
) error {
	const op = "auth.RegisterPending"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate hash", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	err = a.pendingSaver.SavePendingUser(ctx, email, passHash, name, surname, middleName, phoneNumber, birthDate, meta)
	if err != nil {
		// @TODO
	}

	log.Info("added pending user to redis")
	return nil
}

func (a *Auth) ApprovePendingUser(ctx context.Context, userID string) error {
	const op = "auth.ApprovePendingUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("user_id", userID),
	)

	err := a.approver.ApprovePendingUser(ctx, userID)
	if err != nil {
		// @TODO
	}

	log.Info("approved pending user")

	return nil
}

func (a *Auth) GetPendingList(ctx context.Context) ([]models.PendingUser, error) {
	const op = "auth.GetPendingList"

	log := a.log.With(
		slog.String("op", op),
	)

	var users []models.PendingUser

	users, err := a.pendingProvider.ListPendingUsers(ctx)
	if err != nil {
		// @TODO
	}

	log.Info("got pending users")

	return users, err
}

func (a *Auth) DeletePendingUser(ctx context.Context, id string) error {
	const op = "auth.DeletePendingUser"

	log := a.log.With(
		slog.String("op", op),
	)

	err := a.pendingDeleter.DeletePendingUser(ctx, id)
	if err != nil {
		//@TODO
	}

	log.Info("deleted pending user", slog.String("id", id))

	return nil
}

func (a *Auth) GetStudentsList(ctx context.Context) ([]models.Student, error) {
	const op = "auth.GetStudentsList"

	log := a.log.With(
		slog.String("op", op),
	)

	var students []models.Student

	students, err := a.userProvider.StudentsList(ctx)
	if err != nil {
		// @TODO
	}

	log.Info("got students list")

	return students, err
}

func (a *Auth) GetTeachersList(ctx context.Context) ([]models.Teacher, error) {
	const op = "auth.GetTeachersList"

	log := a.log.With(
		slog.String("op", op),
	)

	var teachers []models.Teacher

	teachers, err := a.userProvider.TeacherList(ctx)
	if err != nil {
		// @TODO
	}

	log.Info("got teachers list")

	return teachers, err
}

func (a *Auth) GetAdminsList(ctx context.Context) ([]models.Admin, error) {
	const op = "auth.GetAdminsList"

	log := a.log.With(
		slog.String("op", op),
	)

	var admins []models.Admin

	admins, err := a.userProvider.AdminList(ctx)
	if err != nil {
		// @TODO
	}

	log.Info("got admins list")

	return admins, err
}

func (a *Auth) DeleteUser(ctx context.Context, userID int64, role string) error {
	const op = "auth.DeleteStudent"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("user_id", int(userID)),
	)

	err := a.userDeleter.DeleteUser(ctx, userID, role)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))

			return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		if errors.Is(err, storage.ErrUserDontHasThisRole) {
			log.Warn("user have another role", sl.Err(err))

			return fmt.Errorf("%s: %w", op, ErrUserNotMatchingRole)
		}

		log.Error("failed to delete student", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user  deleted")

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
