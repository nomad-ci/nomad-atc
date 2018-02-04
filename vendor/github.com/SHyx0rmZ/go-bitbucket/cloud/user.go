package cloud

type user struct {
	Name        string `json:"username"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
}

func (u *user) GetName() string {
	return u.Name
}

func (u *user) GetDisplayName() string {
	return u.DisplayName
}
