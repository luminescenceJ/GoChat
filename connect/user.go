package connect

type User struct {
	Id       int    `json:"id"`       // 用户唯一标识，自增设计
	Username string `json:"username"` // 用户名称
	Password string `json:"password"` // 用户密码 密文传输
}
