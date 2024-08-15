package hue

import (
	"github.com/utiiz/go-hue/pkg/bridge"
	"github.com/utiiz/go-hue/pkg/user"
)

type Bridge = bridge.Bridge
type User = user.User

var (
	NewBridge = bridge.NewBridge
	Discover  = bridge.Discover

	NewUser = user.NewUser
)
