package auth

import (
	"context"
	"google.golang.org/grpc"
	protosso "protos/sso"
	"sso/internal/domain/models"
)

type Auth interface {
	Login(ctx context.Context,
		email string,
		password string,
		role string,
	) (token string, err error)
	RegisterStudent(ctx context.Context,
		email string,
		password string,
		name string,
		surname string,
		middleName string,
		phoneNumber string,
		birthDate models.BirthDate,
		group string,
		studentNumber string,
	) (userID int64, err error)
	RegisterTeacher(ctx context.Context,
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
	) (userID int64, err error)
	RegisterAdmin(ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	IsAdmin(ctx context.Context,
		userID int64,
	) (bool, error)
	IsStudent(ctx context.Context,
		userID int64,
	) (bool, error)
	IsTeacher(ctx context.Context,
		userID int64,
	) (bool, error)
}

type serverAPI struct {
	protosso.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	protosso.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) RegisterStudent(ctx context.Context, req *protosso.RegisterStudentRequest) (*protosso.RegisterResponse, error) {
	userID, err := s.auth.RegisterStudent(ctx,
		req.GetEmail(),
		req.GetPassword(),
		req.GetName(),
		req.GetSurname(),
		req.GetPhoneNumber(),
		req.GetMiddleName(),
		models.BirthDate{
			Year:  req.GetBirthDate().GetYear(),
			Month: req.GetBirthDate().GetMonth(),
			Day:   req.GetBirthDate().GetDay(),
		},
		req.GetGroup(),
		req.GetStudentNumber(),
	)
	if err != nil {
		// @Todo implement
	}

	return &protosso.RegisterResponse{UserId: userID}, nil
}

func (s *serverAPI) RegisterTeacher(ctx context.Context, req *protosso.RegisterTeacherRequest) (*protosso.RegisterResponse, error) {
	userID, err := s.auth.RegisterTeacher(ctx,
		req.GetEmail(),
		req.GetPassword(),
		req.GetName(),
		req.GetSurname(),
		req.GetMiddleName(),
		req.GetPhoneNumber(),
		models.BirthDate{
			Year:  req.GetBirthDate().GetYear(),
			Month: req.GetBirthDate().GetMonth(),
			Day:   req.GetBirthDate().GetDay(),
		},
		req.GetTitle(),
		req.GetDepartment(),
		req.GetDegree(),
	)
	if err != nil {
		// @Todo implement
	}

	return &protosso.RegisterResponse{UserId: userID}, nil
}

func (s *serverAPI) RegisterAdmin(ctx context.Context, req *protosso.RegisterAdminRequest) (*protosso.RegisterResponse, error) {
	userID, err := s.auth.RegisterAdmin(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		// @Todo implement
	}

	return &protosso.RegisterResponse{UserId: userID}, nil

}

func (s *serverAPI) Login(ctx context.Context, req *protosso.LoginRequest) (*protosso.LoginResponse, error) {
	tokenTTL, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetRole())
	if err != nil {
		// @Todo implement
	}

	return &protosso.LoginResponse{Token: tokenTTL}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *protosso.IsAdminRequest) (*protosso.IsAdminResponse, error) {
	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		// @Todo implement
	}
	return &protosso.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func (s *serverAPI) IsStudent(ctx context.Context, req *protosso.IsStudentRequest) (*protosso.IsStudentResponse, error) {
	isStudent, err := s.auth.IsStudent(ctx, req.GetUserId())
	if err != nil {
		// @Todo implement
	}
	return &protosso.IsStudentResponse{IsStudent: isStudent}, nil
}

func (s *serverAPI) IsTeacher(ctx context.Context, req *protosso.IsTeacherRequest) (*protosso.IsTeacherResponse, error) {
	isTeacher, err := s.auth.IsTeacher(ctx, req.GetUserId())
	if err != nil {
		// @Todo implement
	}
	return &protosso.IsTeacherResponse{IsTeacher: isTeacher}, nil
}
