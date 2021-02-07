package main

func NewService() Service {
	return service{}
}

type service struct{}

type Service interface {
	HealthCheck() (string, error)
	GetUser(string, string, string) (User, error)
}

type User struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (service) HealthCheck() (string, error) {
	return "ok", nil
}

func (service) GetUser(id string, name string, email string) (User, error) {
	return User{
		Id:    id,
		Name:  name,
		Email: email,
	}, nil
}
