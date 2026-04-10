package game

import gameapi "workflowv2snake/backend/api/game"

type ControllerV1 struct{}

func NewV1() gameapi.IGameV1 {
	return &ControllerV1{}
}
