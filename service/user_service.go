package service

type UsersResponse struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type IUsersService interface {
	List() []UsersResponse
	Get() UsersResponse
	Create() UsersResponse
	Update() UsersResponse
	Delete() UsersResponse
}

type UsersService struct{}

func UsersServiceImpl() IUsersService {
	return &UsersService{}
}

func (u *UsersService) List() []UsersResponse {
	return []UsersResponse{}
}

func (u *UsersService) Get() UsersResponse {
	return UsersResponse{}
}

func (u *UsersService) Create() UsersResponse {
	return UsersResponse{}
}

func (u *UsersService) Update() UsersResponse {
	return UsersResponse{}
}

func (u *UsersService) Delete() UsersResponse {
	return UsersResponse{}
}
