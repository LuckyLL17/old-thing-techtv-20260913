package domain

import "time"

type Category struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Code        string    `gorm:"size:30;uniqueIndex;not null" json:"code"`
	Name        string    `gorm:"size:50;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	Icon        string    `gorm:"size:100" json:"icon"`
	TutorialCount int     `gorm:"default:0" json:"tutorial_count"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (c *Category) TableName() string {
	return "categories"
}

var DefaultCategories = []Category{
	{Code: CatFurniture, Name: "家具", Icon: "🛋️", SortOrder: 1, Description: "旧家具改造翻新"},
	{Code: CatClothing, Name: "服饰", Icon: "👖", SortOrder: 2, Description: "衣物改款再利用"},
	{Code: CatContainer, Name: "容器", Icon: "🫙", SortOrder: 3, Description: "瓶罐盒子改造"},
	{Code: CatDecoration, Name: "装饰", Icon: "🎨", SortOrder: 4, Description: "家居装饰品"},
	{Code: CatTool, Name: "工具", Icon: "🔧", SortOrder: 5, Description: "实用工具制作"},
	{Code: CatElectronic, Name: "电子", Icon: "💡", SortOrder: 6, Description: "电子废物改造"},
	{Code: CatOther, Name: "其他", Icon: "✨", SortOrder: 99, Description: "其他创意改造"},
}
