//go:build windows

package desktop

import "github.com/go-vgo/robotgo"

func NewRobotMouse() *RobotMouse {
	return newRobotMouse(robotgo.Location, robotgo.MoveRelative)
}
