package types

type Bridge interface {
	URL() string
	SetLightOn(id string) error
}
