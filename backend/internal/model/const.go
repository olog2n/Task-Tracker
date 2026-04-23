package model

const (
	TitleMaxLength       int = 100
	TitleMinLength       int = 1
	DescriptionMaxLength int = 1000
	AuthorMinLength      int = 1
)

const (
	StatusBacklog TaskStatus = iota
	StatusInProgress
	StatusReview
	StatusDone
	StatusCancelled
)
