package model

import (
	"github.com/wejectchen/ginblog/utils/errmsg"
	"gorm.io/gorm"
)

// 文章结构体

type Article struct {
	Category Category `gorm:"foreignkey:Cid"`
	gorm.Model
	Title        string `gorm:"type:varchar(100);not null" json:"title"`
	Cid          int    `gorm:"type:int;not null" json:"cid"`
	Desc         string `gorm:"type:varchar(200)" json:"desc"`
	Content      string `gorm:"type:longtext" json:"content"`
	Img          string `gorm:"type:varchar(100)" json:"img"`
	CommentCount int    `gorm:"type:int;not null;default:0" json:"comment_count"`
	ReadCount    int    `gorm:"type:int;not null;default:0" json:"read_count"`
}

// CreateArt 新增文章
func CreateArt(data *Article) int {
	err := db.Create(&data).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// GetCateArt 查询分类下的所有文章
func GetCateArt(id int, pageSize int, pageNum int) ([]Article, int, int64) {
	var cateArtList []Article
	var total int64
	/**
	-- 查询文章列表（分页），并关联查询分类信息
	SELECT
	  articles.*,
	  categories.*  -- 关联分类表的所有字段
	FROM
	  articles
	LEFT JOIN
	  categories ON articles.cid = categories.id  -- 通过 cid 关联分类表
	WHERE
	  articles.cid = 2  -- 筛选分类 ID=2 的文章
	LIMIT 10 OFFSET 0;  -- 取 10 条，跳过 0 条（第 1 页）
	*/
	err = db.Preload("Category").Limit(pageSize).Offset((pageNum-1)*pageSize).Where(
		"cid =?", id).Find(&cateArtList).Error
	/**
	-- 统计分类 ID=2 的文章总数
	SELECT COUNT(*) FROM articles WHERE cid = 2;
	*/
	db.Model(&cateArtList).Where("cid =?", id).Count(&total)
	if err != nil {
		return nil, errmsg.ERROR_CATE_NOT_EXIST, 0
	}
	return cateArtList, errmsg.SUCCESS, total
}

// GetArtInfo 查询单个文章 查询单篇文章的详细信息，并自动将该文章的阅读量 +1
func GetArtInfo(id int) (Article, int) {
	var art Article
	err = db.Where("id = ?", id).Preload("Category").First(&art).Error

	db.Model(&art).Where("id = ?", id).UpdateColumn("read_count", gorm.Expr("read_count + ?", 1))
	if err != nil {
		return art, errmsg.ERROR_ART_NOT_EXIST
	}
	return art, errmsg.SUCCESS
}

// GetArt 查询文章列表
func GetArt(pageSize int, pageNum int) ([]Article, int, int64) {
	var articleList []Article
	var err error
	var total int64
	/**
	SELECT
	  article.id,
	  title,
	  img,
	  created_at,
	  updated_at,
	  `desc`,
	  comment_count,
	  read_count,
	  category.name
	FROM
	  articles
	INNER JOIN
	  categories ON articles.cid = categories.id  -- 关联分类表（通过cid）
	ORDER BY
	  created_at DESC  -- 按创建时间倒序（最新的在前）
	LIMIT 10 OFFSET 0;  -- 取10条，跳过0条（第1页）
	*/
	err = db.Select("article.id, title, img, created_at, updated_at, `desc`, comment_count, read_count, category.name").Limit(pageSize).Offset((pageNum - 1) * pageSize).Order("Created_At DESC").Joins("Category").Find(&articleList).Error
	// 单独计数	SELECT COUNT(*) FROM articles;
	db.Model(&articleList).Count(&total)
	if err != nil {
		return nil, errmsg.ERROR, 0
	}
	return articleList, errmsg.SUCCESS, total

}

// SearchArticle 搜索文章标题
func SearchArticle(title string, pageSize int, pageNum int) ([]Article, int, int64) {
	var articleList []Article
	var err error
	var total int64
	/**
	SELECT
	  article.id,
	  title,
	  img,
	  created_at,
	  updated_at,
	  `desc`,
	  comment_count,
	  read_count,
	  Category.name
	FROM
	  articles
	INNER JOIN
	  categories ON articles.cid = categories.id  -- 关联分类表，获取分类名称
	WHERE
	  title LIKE 'go%'  -- 模糊匹配：标题以"go"开头（如"golang"、"go入门"等）
	ORDER BY
	  created_at DESC  -- 按创建时间倒序（最新的在前）
	LIMIT 10 OFFSET 0;  -- 取10条，跳过0条（第1页）
	*/
	err = db.Select("article.id,title, img, created_at, updated_at, `desc`, comment_count, read_count, Category.name").Order("Created_At DESC").Joins("Category").Where("title LIKE ?",
		title+"%",
	).Limit(pageSize).Offset((pageNum - 1) * pageSize).Find(&articleList).Error
	//单独计数 SELECT COUNT(*) FROM articles WHERE title LIKE 'go%';  -- 同样的模糊匹配条件
	db.Model(&articleList).Where("title LIKE ?",
		title+"%",
	).Count(&total)

	if err != nil {
		return nil, errmsg.ERROR, 0
	}
	return articleList, errmsg.SUCCESS, total
}

// EditArt 编辑文章
func EditArt(id int, data *Article) int {
	var art Article
	var maps = make(map[string]interface{})
	maps["title"] = data.Title
	maps["cid"] = data.Cid
	maps["desc"] = data.Desc
	maps["content"] = data.Content
	maps["img"] = data.Img
	/**
	-- 根据 ID 更新文章的指定字段
	UPDATE articles
	SET
	  title = '新标题',
	  cid = 3,
	  `desc` = '新描述',
	  content = '新内容',
	  img = 'new-img.png',
	  updated_at = '当前时间'  -- GORM 自动更新 updated_at 字段
	WHERE
	  id = 5;  -- 只更新 ID=5 的文章
	*/
	err = db.Model(&art).Where("id = ? ", id).Updates(&maps).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// DeleteArt 删除文章
func DeleteArt(id int) int {
	var art Article
	/**
	-- 软删除 ID=5 的文章（GORM 默认行为）
	UPDATE articles
	SET deleted_at = '当前时间戳'  -- 写入删除时间，标记为已删除
	WHERE id = 5 AND deleted_at IS NULL;  -- 只删除未被标记删除的记录
	*/
	err = db.Where("id = ? ", id).Delete(&art).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}
