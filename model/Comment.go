package model

import (
	"github.com/wejectchen/ginblog/utils/errmsg"
	"gorm.io/gorm"
)

// Comment 评论结构体
type Comment struct {
	gorm.Model
	UserId    uint   `json:"user_id"`
	ArticleId uint   `json:"article_id"`
	Title     string `gorm:"type:varchar(500);not null;" json:"article_title"`
	Username  string `gorm:"type:varchar(500);not null;" json:"username"`
	Content   string `gorm:"type:varchar(500);not null;" json:"content"`
	Status    int8   `gorm:"type:tinyint;default:2" json:"status"`
}

// AddComment 新增评论
func AddComment(data *Comment) int {
	err = db.Create(&data).Error
	/**
	-- 假设新增评论的 user_id=10、article_id=5、content="很棒的文章"、status=0（待审核）
	INSERT INTO comments (user_id, article_id, content, status, created_at, updated_at)
	VALUES (10, 5, '很棒的文章', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
	*/
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// GetComment 查询单个评论
func GetComment(id int) (Comment, int) {
	var comment Comment
	// -- 假设查询 ID=3 的评论
	//SELECT * FROM comments WHERE id = 3 LIMIT 1;
	err = db.Where("id = ?", id).First(&comment).Error
	if err != nil {
		return comment, errmsg.ERROR
	}
	return comment, errmsg.SUCCESS
}

// GetCommentList 后台所有获取评论列表
func GetCommentList(pageSize int, pageNum int) ([]Comment, int64, int) {

	var commentList []Comment
	var total int64
	//-- 1. 统计总条数（所有评论）
	//SELECT COUNT(*) AS total FROM comments;
	db.Find(&commentList).Count(&total)
	/**
	-- 2. 分页查询（假设 pageSize=10，pageNum=1）
	SELECT
	  comments.id,
	  articles.title,
	  comments.user_id,
	  comments.article_id,
	  users.username,
	  comments.content,
	  comments.status,
	  comments.created_at,
	  comments.deleted_at
	FROM comments
	LEFT JOIN articles ON comments.article_id = articles.id  -- 关联文章表
	LEFT JOIN users ON comments.user_id = users.id  -- 关联用户表
	ORDER BY comments.created_at DESC  -- 按创建时间倒序
	LIMIT 10 OFFSET 0;  -- 分页：每页10条，第1页（OFFSET=(1-1)*10=0）
	*/
	err = db.Model(&commentList).Limit(pageSize).Offset((pageNum - 1) * pageSize).Order("Created_At DESC").Select("comment.id, article.title,user_id,article_id, user.username, comment.content, comment.status,comment.created_at,comment.deleted_at").Joins("LEFT JOIN article ON comment.article_id = article.id").Joins("LEFT JOIN user ON comment.user_id = user.id").Scan(&commentList).Error
	if err != nil {
		return commentList, 0, errmsg.ERROR
	}
	return commentList, total, errmsg.SUCCESS
}

// GetCommentCount 获取评论数量
func GetCommentCount(id int) int64 {
	var comment Comment
	var total int64
	/**
	-- 假设统计 article_id=5 且 status=1 的评论数
	SELECT COUNT(*) AS total FROM comments
	WHERE article_id = 5 AND status = 1;
	*/
	db.Find(&comment).Where("article_id = ?", id).Where("status = ?", 1).Count(&total)
	return total
}

// GetCommentListFront 展示页面获取评论列表
func GetCommentListFront(id int, pageSize int, pageNum int) ([]Comment, int64, int) {
	var commentList []Comment
	var total int64
	/**
	-- 1. 统计总条数（指定文章+已通过）
	SELECT COUNT(*) AS total FROM comments
	WHERE article_id = 5 AND status = 1;  -- 假设 article_id=5
	*/
	db.Find(&Comment{}).Where("article_id = ?", id).Where("status = ?", 1).Count(&total)
	/**
	-- 2. 分页查询（假设 pageSize=10，pageNum=1）
	SELECT
	  comments.id,
	  articles.title,
	  comments.user_id,
	  comments.article_id,
	  users.username,
	  comments.content,
	  comments.status,
	  comments.created_at,
	  comments.deleted_at
	FROM comments
	LEFT JOIN articles ON comments.article_id = articles.id
	LEFT JOIN users ON comments.user_id = users.id
	WHERE comments.article_id = 5 AND comments.status = 1  -- 条件：指定文章+已通过
	ORDER BY comments.created_at DESC
	LIMIT 10 OFFSET 0;
	*/
	err = db.Model(&Comment{}).Limit(pageSize).Offset((pageNum-1)*pageSize).Order("Created_At DESC").Select("comment.id, article.title, user_id, article_id, user.username, comment.content, comment.status,comment.created_at,comment.deleted_at").Joins("LEFT JOIN article ON comment.article_id = article.id").Joins("LEFT JOIN user ON comment.user_id = user.id").Where("article_id = ?",
		id).Where("status = ?", 1).Scan(&commentList).Error
	if err != nil {
		return commentList, 0, errmsg.ERROR
	}
	return commentList, total, errmsg.SUCCESS
}

// 编辑评论（暂不允许编辑评论）

// DeleteComment 删除评论
func DeleteComment(id uint) int {
	var comment Comment
	/**
	-- 软删除（逻辑删除，假设评论 ID=8）
	UPDATE comments
	SET deleted_at = CURRENT_TIMESTAMP
	WHERE id = 8;

	-- 若开启物理删除（使用 Unscoped()），则为：
	DELETE FROM comments WHERE id = 8;
	*/
	err = db.Where("id = ?", id).Delete(&comment).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// CheckComment 通过评论
func CheckComment(id int, data *Comment) int {
	var comment Comment
	var res Comment
	var article Article
	var maps = make(map[string]interface{})
	maps["status"] = data.Status
	/**
	-- 先更新评论状态
	UPDATE comments
	SET status = 1, updated_at = CURRENT_TIMESTAMP  -- GORM 自动更新更新时间
	WHERE id = 5;

	-- 再查询更新后的完整评论信息（用于获取关联的文章 ID）
	SELECT * FROM comments WHERE id = 5 LIMIT 1;
	*/
	err = db.Model(&comment).Where("id = ?", id).Updates(maps).First(&res).Error
	/**
	 文章评论数 +1（对应文章 ID=5）
	UPDATE articles
	SET comment_count = comment_count + 1
	WHERE id = 5;
	*/
	db.Model(&article).Where("id = ?", res.ArticleId).UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1))
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// UncheckComment 撤下评论
func UncheckComment(id int, data *Comment) int {
	var comment Comment
	var res Comment
	var article Article
	var maps = make(map[string]interface{})
	maps["status"] = data.Status

	/**
	-- 1. 更新评论状态（假设评论 ID=6，更新 status=0）
	UPDATE comments
	SET status = 0, updated_at = CURRENT_TIMESTAMP
	WHERE id = 6;

	-- 2. 关联查询该评论对应的文章 ID（假设查询到 article_id=5）
	SELECT article_id FROM comments WHERE id = 6 LIMIT 1;

	-- 3. 文章评论数 -1（对应文章 ID=5）
	UPDATE articles
	SET comment_count = comment_count - 1
	WHERE id = 5;
	*/
	err = db.Model(&comment).Where("id = ?", id).Updates(maps).First(&res).Error
	db.Model(&article).Where("id = ?", res.ArticleId).UpdateColumn("comment_count", gorm.Expr("comment_count - ?", 1))
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}
