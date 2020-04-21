package models

type Model interface {
	Marshal() ([]byte, error)
}
