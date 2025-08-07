package model

import (
	"github.com/wejectchen/ginblog/utils/errmsg"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
)

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(20);not null " json:"username" validate:"required,min=4,max=12" label:"用户名"`
	Password string `gorm:"type:varchar(500);not null" json:"password" validate:"required,min=6,max=120" label:"密码"`
	Role     int    `gorm:"type:int;DEFAULT:2" json:"role" validate:"required,gte=2" label:"角色码"`
}

// CheckUser 查询用户是否存在
func CheckUser(name string) (code int) {
	var user User
	/**
	-- 假设检查用户名为 "user"
	SELECT id FROM users WHERE username = 'user' LIMIT 1;
	*/
	db.Select("id").Where("username = ?", name).First(&user)
	if user.ID > 0 {
		return errmsg.ERROR_USERNAME_USED //1001
	}
	return errmsg.SUCCESS
}

// CheckUpUser 更新查询
func CheckUpUser(id int, name string) (code int) {
	var user User
	/**
	-- 假设用户 ID=5，检查更新用户名 "name"
	SELECT id, username FROM user WHERE username = 'name' LIMIT 1;
	*/
	db.Select("id, username").Where("username = ?", name).First(&user)
	if user.ID == uint(id) {
		return errmsg.SUCCESS
	}
	if user.ID > 0 {
		return errmsg.ERROR_USERNAME_USED //1001
	}
	return errmsg.SUCCESS
}

// CreateUser 新增用户
func CreateUser(data *User) int {
	//data.Password = ScryptPw(data.Password)
	/**
	-- 假设新增用户：username="newuser"，password="加密后的密码"，role=2（默认普通用户）
	INSERT INTO user (username, password, role, created_at, updated_at)
	VALUES ('newuser', '$2a$10$加密后的密码', 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
	*/
	err := db.Create(&data).Error
	if err != nil {
		return errmsg.ERROR // 500
	}
	return errmsg.SUCCESS
}

// GetUser 查询用户
func GetUser(id int) (User, int) {
	var user User
	//-- 假设查询 ID=5 的用户
	//SELECT * FROM users WHERE ID = 5 LIMIT 1;
	err := db.Limit(1).Where("ID = ?", id).Find(&user).Error
	if err != nil {
		return user, errmsg.ERROR
	}
	return user, errmsg.SUCCESS
}

// GetUsers 查询用户列表
func GetUsers(username string, pageSize int, pageNum int) ([]User, int64) {
	var users []User
	var total int64
	/**
	-- 场景1：带用户名模糊搜索（如 username="test"，pageSize=10，pageNum=1）
	SELECT id, username, role, created_at
	FROM users
	WHERE username LIKE 'test%'  -- 匹配以 "test" 开头的用户名
	LIMIT 10 OFFSET 0;  -- 分页：第1页，每页10条

	-- 统计符合条件的总条数
	SELECT COUNT(*) AS total FROM user WHERE username LIKE 'test%';
	*/
	if username != "" {
		db.Select("id,username,role,created_at").Where(
			"username LIKE ?", username+"%",
		).Limit(pageSize).Offset((pageNum - 1) * pageSize).Find(&users)
		db.Model(&users).Where(
			"username LIKE ?", username+"%",
		).Count(&total)
		return users, total
	}
	/**
	-- 场景2：无搜索条件（查询所有用户）
	SELECT id, username, role, created_at
	FROM user
	LIMIT 10 OFFSET 0;

	-- 统计总条数
	SELECT COUNT(*) AS total FROM users;
	*/
	db.Select("id,username,role,created_at").Limit(pageSize).Offset((pageNum - 1) * pageSize).Find(&users)
	db.Model(&users).Count(&total)

	if err != nil {
		return users, 0
	}
	return users, total
}

// EditUser 编辑用户信息
func EditUser(id int, data *User) int {
	var user User
	var maps = make(map[string]interface{})
	maps["username"] = data.Username
	maps["role"] = data.Role
	/**
	-- 假设更新 ID=5 的用户：username="updateduser"，role=1（管理员）
	UPDATE user
	SET username = 'updateduser', role = 1, updated_at = CURRENT_TIMESTAMP
	WHERE id = 5;
	*/
	err = db.Model(&user).Where("id = ? ", id).Updates(maps).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// ChangePassword 修改密码
func ChangePassword(id int, data *User) int {
	//var user User
	//var maps = make(map[string]interface{})
	//maps["password"] = data.Password
	/**
	-- 假设更新 ID=5 的用户密码（新密码已加密）
	UPDATE user
	SET password = '$2a$10$新加密的密码', updated_at = CURRENT_TIMESTAMP
	WHERE id = 5;
	*/
	err = db.Select("password").Where("id = ?", id).Updates(&data).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// DeleteUser 删除用户
func DeleteUser(id int) int {
	var user User
	/**
	-- 软删除（逻辑删除，标记 deleted_at）
	UPDATE user SET deleted_at = CURRENT_TIMESTAMP WHERE id = 5;

	-- 若开启物理删除（使用 Unscoped()）
	DELETE FROM user WHERE id = 5;
	*/
	err = db.Where("id = ? ", id).Delete(&user).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// BeforeCreate 密码加密&权限控制
func (u *User) BeforeCreate(_ *gorm.DB) (err error) {
	u.Password = ScryptPw(u.Password)
	u.Role = 2
	return nil
}

func (u *User) BeforeUpdate(_ *gorm.DB) (err error) {
	u.Password = ScryptPw(u.Password)
	return nil
}

// ScryptPw 生成密码
func ScryptPw(password string) string {
	const cost = 10

	HashPw, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		log.Fatal(err)
	}

	return string(HashPw)
}

// CheckLogin 后台登录验证
func CheckLogin(username string, password string) (User, int) {
	var user User
	var PasswordErr error

	/**
	-- 1. 查询用户名对应的用户信息（假设用户名为 "admin"）
	SELECT * FROM users WHERE username = 'admin' LIMIT 1;
	*/
	db.Where("username = ?", username).First(&user)

	PasswordErr = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if user.ID == 0 {
		return user, errmsg.ERROR_USER_NOT_EXIST
	}
	if PasswordErr != nil {
		return user, errmsg.ERROR_PASSWORD_WRONG
	}
	if user.Role != 1 {
		return user, errmsg.ERROR_USER_NO_RIGHT
	}
	return user, errmsg.SUCCESS
}

// CheckLoginFront 前台登录
func CheckLoginFront(username string, password string) (User, int) {
	var user User
	var PasswordErr error

	db.Where("username = ?", username).First(&user)

	PasswordErr = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if user.ID == 0 {
		return user, errmsg.ERROR_USER_NOT_EXIST
	}
	if PasswordErr != nil {
		return user, errmsg.ERROR_PASSWORD_WRONG
	}
	return user, errmsg.SUCCESS
}
