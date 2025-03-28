package dto

type Request interface {
	Validate() error
}
