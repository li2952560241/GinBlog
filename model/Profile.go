package model

import (
	"github.com/wejectchen/ginblog/utils/errmsg"
)

type Profile struct {
	ID        int    `gorm:"primaryKey" json:"id"`
	Name      string `gorm:"type:varchar(20)" json:"name"`
	Desc      string `gorm:"type:varchar(200)" json:"desc"`
	Qqchat    string `gorm:"type:varchar(200)" json:"qq_chat"`
	Wechat    string `gorm:"type:varchar(100)" json:"wechat"`
	Weibo     string `gorm:"type:varchar(200)" json:"weibo"`
	Bili      string `gorm:"type:varchar(200)" json:"bili"`
	Email     string `gorm:"type:varchar(200)" json:"email"`
	Img       string `gorm:"type:varchar(200)" json:"img"`
	Avatar    string `gorm:"type:varchar(200)" json:"avatar"`
	IcpRecord string `gorm:"type:varchar(200)" json:"icp_record"`
}

// GetProfile 获取个人信息设置
func GetProfile(id int) (Profile, int) {
	var profile Profile
	//-- 假设要查询 ID=10 的用户个人信息
	//SELECT * FROM profiles WHERE ID = 10 LIMIT 1;
	err = db.Where("ID = ?", id).First(&profile).Error
	if err != nil {
		return profile, errmsg.ERROR
	}
	return profile, errmsg.SUCCESS
}

// UpdateProfile 更新个人信息设置
func UpdateProfile(id int, data *Profile) int {
	var profile Profile
	/** -- 假设更新 ID=10 的用户信息，新数据为：nickname="新昵称", avatar="new_avatar.png"
	UPDATE profiles
	SET
	  nickname = '新昵称',
	  avatar = 'new_avatar.png',
	  -- 其他需要更新的字段（如简介、性别等，根据 data 中的字段动态生成）
	  updated_at = CURRENT_TIMESTAMP  -- GORM 自动更新更新时间
	WHERE ID = 10;
	*/
	err = db.Model(&profile).Where("ID = ?", id).Updates(&data).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}
