package response

import (
	protosso "protos/sso"
	"sso/internal/domain/models"
)

func StudentsToGRPC(students []models.Student) []*protosso.StudentBody {
	var result []*protosso.StudentBody
	for _, s := range students {
		result = append(result, &protosso.StudentBody{
			UserInfo: &protosso.UserBody{
				Email:       s.Email,
				Name:        s.Name,
				Surname:     s.Surname,
				MiddleName:  s.MiddleName,
				PhoneNumber: s.PhoneNumber,
				BirthDate: &protosso.Date{
					Year:  s.BirthDate.Year,
					Month: s.BirthDate.Month,
					Day:   s.BirthDate.Day,
				},
			},
			Group:         s.Group,
			StudentNumber: s.StudentNumber,
		})
	}
	return result
}

func TeachersToGRPC(teachers []models.Teacher) []*protosso.TeacherBody {
	var result []*protosso.TeacherBody
	for _, t := range teachers {
		result = append(result, &protosso.TeacherBody{
			UserInfo: &protosso.UserBody{
				Email:       t.Email,
				Name:        t.Name,
				Surname:     t.Surname,
				MiddleName:  t.MiddleName,
				PhoneNumber: t.PhoneNumber,
				BirthDate: &protosso.Date{
					Year:  t.BirthDate.Year,
					Month: t.BirthDate.Month,
					Day:   t.BirthDate.Day,
				},
			},
			Title:      t.Title,
			Department: t.Department,
			Degree:     t.Degree,
		})
	}
	return result
}

func AdminsToGRPC(admins []models.Admin) []*protosso.AdminBody {
	var result []*protosso.AdminBody
	for _, a := range admins {
		result = append(result, &protosso.AdminBody{
			Email: a.Email,
			// Пропускаем пароль
		})
	}
	return result
}

func PendingUsersToGRPC(users []models.PendingUser) []*protosso.PendingUser {
	var result []*protosso.PendingUser
	for _, u := range users {
		result = append(result, &protosso.PendingUser{
			UserInfo: &protosso.UserBody{
				Email:       u.Email,
				Name:        u.Name,
				Surname:     u.Surname,
				MiddleName:  u.MiddleName,
				PhoneNumber: u.PhoneNumber,
				BirthDate: &protosso.Date{
					Year:  u.BirthDate.Year,
					Month: u.BirthDate.Month,
					Day:   u.BirthDate.Day,
				},
			},
			Role: u.Role,
			Meta: u.Meta,
		})
	}
	return result
}
