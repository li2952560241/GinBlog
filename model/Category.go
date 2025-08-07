package model

import (
	"github.com/wejectchen/ginblog/utils/errmsg"
	"gorm.io/gorm"
)

//文章分类的结构表

type Category struct {
	ID   uint   `gorm:"primary_key;auto_increment" json:"id"`
	Name string `gorm:"type:varchar(20);not null" json:"name"`
}

// CheckCategory 查询分类是否存在
// @name 传过来的name字符串
func CheckCategory(name string) (code int) {
	var cate Category
	//-- 查询名称为"技术"的分类的ID  SELECT id FROM categories WHERE name = '技术' LIMIT 1;
	db.Select("id").Where("name = ?", name).First(&cate)
	if cate.ID > 0 {
		return errmsg.ERROR_CATENAME_USED
	}
	return errmsg.SUCCESS
}

// CreateCate 新增分类
func CreateCate(data *Category) int {
	/**-- 向分类表插入一条新记录
	INSERT INTO categories (name, created_at, updated_at)
	VALUES ('后端开发', '当前时间戳', '当前时间戳');
	*/
	err := db.Create(&data).Error
	if err != nil {
		return errmsg.ERROR // 500
	}
	return errmsg.SUCCESS
}

// GetCateInfo 查询单个分类信息
func GetCateInfo(id int) (Category, int) {
	var cate Category
	//-- 查询ID为3的分类信息
	//SELECT * FROM categories WHERE id = 3 LIMIT 1;
	db.Where("id = ?", id).First(&cate)
	return cate, errmsg.SUCCESS
}

// GetCate 查询分类列表
func GetCate(pageSize int, pageNum int) ([]Category, int64) {
	var cate []Category
	var total int64
	/**
	-- 1. 查询分页数据（假设 pageSize=10，pageNum=1）
	SELECT * FROM categories
	LIMIT 10 OFFSET 0;  -- OFFSET 计算方式：(pageNum-1)*pageSize
	*/
	err = db.Find(&cate).Limit(pageSize).Offset((pageNum - 1) * pageSize).Error
	/**
	-- 2. 统计总条数
	SELECT COUNT(*) AS total FROM categories;
	*/
	db.Model(&cate).Count(&total)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0
	}
	return cate, total
}

// EditCate 编辑分类信息
func EditCate(id int, data *Category) int {
	var cate Category
	var maps = make(map[string]interface{})
	maps["name"] = data.Name
	/**
	-- 假设 id=5，要更新的分类名称为 "新分类名"
	UPDATE categories
	SET name = '新分类名', updated_at = CURRENT_TIMESTAMP  -- GORM 自动更新更新时间
	WHERE id = 5;
	*/
	err = db.Model(&cate).Where("id = ? ", id).Updates(maps).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// DeleteCate 删除分类
func DeleteCate(id int) int {
	var cate Category
	/**
	-- 软删除（逻辑删除，仅更新 deleted_at 字段）
	-- 假设 id=5
	UPDATE categories
	SET deleted_at = CURRENT_TIMESTAMP
	WHERE id = 5;

	-- 若开启了物理删除（使用 Unscoped()），则对应：
	DELETE FROM categories WHERE id = 5;
	*/
	err = db.Where("id = ? ", id).Delete(&cate).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}
