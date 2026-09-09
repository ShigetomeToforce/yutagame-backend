package model

import "time"

type ContactInquiry struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	Email     string    `gorm:"column:email;not null" json:"email"`
	Subject   string    `gorm:"column:subject;not null" json:"subject"`
	Message   string    `gorm:"column:message;not null" json:"message"`
	Status    string    `gorm:"column:status;not null" json:"status"`
	AdminNote string    `gorm:"column:admin_note;not null" json:"adminNote"`
	CreatedAt time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`
}
