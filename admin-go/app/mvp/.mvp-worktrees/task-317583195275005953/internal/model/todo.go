package model

import "time"

// Todo 表示待办事项实体
type Todo struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTodoRequest 创建待办事项请求
type CreateTodoRequest struct {
	Title string `json:"title"`
}

// CreateTodoResponse 创建待办事项响应
type CreateTodoResponse struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

// UpdateTodoRequest 更新待办事项请求
type UpdateTodoRequest struct {
	Title     string `json:"title"`
	Completed *bool  `json:"completed,omitempty"`
}

// UpdateTodoResponse 更新待办事项响应
type UpdateTodoResponse struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

// GetTodoResponse 获取单个待办事项响应
type GetTodoResponse struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

// ListTodoResponse 获取待办事项列表响应
type ListTodoResponse struct {
	Todos []GetTodoResponse `json:"todos"`
	Total int64             `json:"total"`
}

// DeleteTodoResponse 删除待办事项响应
type DeleteTodoResponse struct {
	Success bool `json:"success"`
}
