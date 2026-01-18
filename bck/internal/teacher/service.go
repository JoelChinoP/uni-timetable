package teacher

import (
	"context"
	"time"

	db "github.com/JoelChinoP/timetable_bck/internal/database"
	sqlc "github.com/JoelChinoP/timetable_bck/internal/database/sqlc"
)

type TeacherService interface {
	CreateTeacher(ctx context.Context, t *CreateTeacherDTO) (*TeacherDTO, error)
	GetTeacherByID(ctx context.Context, id int32) (*TeacherDTO, error)
	ListTeachers(ctx context.Context) ([]*TeacherDTO, error)
	UpdateTeacher(ctx context.Context, id int32, t *UpdateTeacherDTO) (*TeacherDTO, error)
	DeleteTeacher(ctx context.Context, id int32) error
}

type service struct {
	store *sqlc.Queries
}

// GetTeacherService crea una nueva instancia del servicio de profesores
func GetTeacherService() TeacherService {
	return &service{
		store: sqlc.New(db.Pool()),
	}
}

func (serv *service) CreateTeacher(ctx context.Context, t *CreateTeacherDTO) (*TeacherDTO, error) {
	params := sqlc.CreateTeacherParams{
		Name:     t.Name,
		LastName: t.LastName,
	}

	teacher, err := serv.store.CreateTeacher(ctx, params)
	if err != nil {
		return nil, err
	}
	return mapperToDTO(teacher), nil
}

func (serv *service) GetTeacherByID(ctx context.Context, id int32) (*TeacherDTO, error) {
	teacher, err := serv.store.GetTeacher(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapperToDTO(teacher), nil
}

func (serv *service) ListTeachers(ctx context.Context) ([]*TeacherDTO, error) {
	teachers, err := serv.store.ListTeachers(ctx)
	if err != nil {
		return nil, err
	}
	var result []*TeacherDTO
	for _, teacher := range teachers {
		result = append(result, mapperToDTO(teacher))
	}
	return result, nil
}

func (serv *service) UpdateTeacher(ctx context.Context, id int32, t *UpdateTeacherDTO) (*TeacherDTO, error) {
	params := sqlc.UpdateTeacherParams{
		ID:       id,
		Name:     t.Name,
		LastName: t.LastName,
	}
	teacher, err := serv.store.UpdateTeacher(ctx, params)
	if err != nil {
		return nil, err
	}
	return mapperToDTO(teacher), nil
}

func (serv *service) DeleteTeacher(ctx context.Context, id int32) error {
	return serv.store.DeleteTeacher(ctx, id)
}

func mapperToDTO(teacher sqlc.Teacher) *TeacherDTO {
	var createdAt time.Time
	if teacher.CreatedAt.Valid {
		createdAt = teacher.CreatedAt.Time
	}

	return &TeacherDTO{
		ID:        teacher.ID,
		Name:      teacher.Name,
		LastName:  teacher.LastName,
		CreatedAt: createdAt,
	}
}
