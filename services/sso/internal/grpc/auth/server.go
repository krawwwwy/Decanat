package auth

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	protosso "protos/sso"
	"sso/internal/domain/models"
	"sso/internal/lib/response"
	"sso/internal/service/auth"
)

type Auth interface {
	Login(ctx context.Context,
		email string,
		password string,
		role string,
	) (token string, err error)

	RegisterPending(ctx context.Context,
		email string,
		password string,
		name string,
		surname string,
		middleName string,
		phoneNumber string,
		birthDate models.BirthDate,
		role string,
		meta map[string]string,
	) error

	GetPendingList(ctx context.Context) ([]models.PendingUser, error)

	DeletePendingUser(ctx context.Context, userID string) error

	ApprovePendingUser(ctx context.Context, userID string) error

	GetStudentsList(context.Context) ([]models.Student, error)
	GetTeachersList(context.Context) ([]models.Teacher, error)
	GetAdminsList(context.Context) ([]models.Admin, error)

	DeleteUser(ctx context.Context, userID int64, role string) (err error)

	IsAdmin(ctx context.Context, userID int64) (bool, error)
	IsStudent(ctx context.Context, userID int64) (bool, error)
	IsTeacher(ctx context.Context, userID int64) (bool, error)
}

type serverAPI struct {
	protosso.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	protosso.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) RegisterPending(ctx context.Context, req *protosso.RegisterPendingRequest) (*protosso.RegisterPendingResponse, error) {
	err := s.auth.RegisterPending(ctx,
		req.GetInfo().GetUserInfo().GetEmail(),
		req.GetInfo().GetUserInfo().GetPassword(),
		req.GetInfo().GetUserInfo().GetName(),
		req.GetInfo().GetUserInfo().GetSurname(),
		req.GetInfo().GetUserInfo().GetPhoneNumber(),
		req.GetInfo().GetUserInfo().GetMiddleName(),
		models.BirthDate{
			Year:  req.GetInfo().GetUserInfo().GetBirthDate().GetYear(),
			Month: req.GetInfo().GetUserInfo().GetBirthDate().GetMonth(),
			Day:   req.GetInfo().GetUserInfo().GetBirthDate().GetDay(),
		},
		req.GetInfo().GetRole(),
		req.GetInfo().GetMeta(),
	)
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {

			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}
	return nil, nil
}

func (s *serverAPI) GetPendingList(ctx context.Context, _ *emptypb.Empty) (*protosso.GetPendingListResponse, error) {
	var users []models.PendingUser
	users, err := s.auth.GetPendingList(ctx)
	if err != nil {
		// @TODO
	}

	usersGRPC := response.PendingUsersToGRPC(users)
	return &protosso.GetPendingListResponse{Users: usersGRPC}, nil
}

func (s *serverAPI) ApprovePendingUser(ctx context.Context, req *protosso.ApprovePendingUserRequest) (*emptypb.Empty, error) {
	err := s.auth.ApprovePendingUser(ctx, req.GetUserId())
	if err != nil {
		// @TODO
	}

	return nil, nil
}

func (s *serverAPI) DeletePendingUser(ctx context.Context, req *protosso.DeletePendingUserRequest) (*emptypb.Empty, error) {
	err := s.auth.DeletePendingUser(ctx, req.GetUserId())
	if err != nil {
		// @TODO
	}

	return nil, nil
}

func (s *serverAPI) Login(ctx context.Context, req *protosso.LoginRequest) (*protosso.LoginResponse, error) {
	tokenTTL, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetRole())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {

			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		if errors.Is(err, auth.ErrUserNotMatchingRole) {

			return nil, status.Error(codes.NotFound, "you have an account with another role")
		}
		if errors.Is(err, auth.ErrRoleNotExists) {

			return nil, status.Error(codes.FailedPrecondition, "role doesn't exist")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &protosso.LoginResponse{Token: tokenTTL}, nil
}

func (s *serverAPI) GetStudentsList(ctx context.Context, _ *emptypb.Empty) (*protosso.GetStudentsListResponse, error) {
	var students []models.Student

	students, err := s.auth.GetStudentsList(ctx)
	if err != nil {
		//@TODO
	}

	studentGRPC := response.StudentsToGRPC(students)
	return &protosso.GetStudentsListResponse{Students: studentGRPC}, nil
}

func (s *serverAPI) GetTeachersList(ctx context.Context, _ *emptypb.Empty) (*protosso.GetTeachersListResponse, error) {
	var teachers []models.Teacher

	teachers, err := s.auth.GetTeachersList(ctx)
	if err != nil {
		//@TODO
	}

	teachersGRPC := response.TeachersToGRPC(teachers)
	return &protosso.GetTeachersListResponse{Teachers: teachersGRPC}, nil
}

func (s *serverAPI) GetAdminsList(ctx context.Context, _ *emptypb.Empty) (*protosso.GetAdminsListResponse, error) {
	var admins []models.Admin

	admins, err := s.auth.GetAdminsList(ctx)
	if err != nil {
		//@TODO
	}

	adminsGRPC := response.AdminsToGRPC(admins)
	return &protosso.GetAdminsListResponse{Admins: adminsGRPC}, nil
}

func (s *serverAPI) DeleteUser(ctx context.Context, req *protosso.DeleteUserRequest) (*emptypb.Empty, error) {
	err := s.auth.DeleteUser(ctx, req.GetUserId(), req.GetRole())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {

			return nil, status.Error(codes.InvalidArgument, "user not found")
		}
		if errors.Is(err, auth.ErrUserNotMatchingRole) {

			return nil, status.Error(codes.InvalidArgument, "user dont have an account with this role")
		}

		return nil, status.Error(codes.Internal, "internal server error")

	}
	return nil, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *protosso.IsAdminRequest) (*protosso.IsAdminResponse, error) {
	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidUserID) {

			return nil, status.Error(codes.InvalidArgument, "user not found")
		}

		return nil, status.Error(codes.Internal, "internal server error")

	}
	return &protosso.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func (s *serverAPI) IsStudent(ctx context.Context, req *protosso.IsStudentRequest) (*protosso.IsStudentResponse, error) {
	isStudent, err := s.auth.IsStudent(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidUserID) {

			return nil, status.Error(codes.InvalidArgument, "user not found")
		}

		return nil, status.Error(codes.Internal, "internal server error")

	}
	return &protosso.IsStudentResponse{IsStudent: isStudent}, nil
}

func (s *serverAPI) IsTeacher(ctx context.Context, req *protosso.IsTeacherRequest) (*protosso.IsTeacherResponse, error) {
	isTeacher, err := s.auth.IsTeacher(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidUserID) {

			return nil, status.Error(codes.InvalidArgument, "user not found")
		}

		return nil, status.Error(codes.Internal, "internal server error")

	}
	return &protosso.IsTeacherResponse{IsTeacher: isTeacher}, nil
}
