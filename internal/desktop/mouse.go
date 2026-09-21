package desktop

import (
	"fmt"

	"mousekeeper/internal/keeper"
)

type RobotMouse struct {
	location func() (int, int)
	move     func(int, int)
}

func newRobotMouse(location func() (int, int), move func(int, int)) *RobotMouse {
	return &RobotMouse{location: location, move: move}
}

func (m *RobotMouse) Position() (point keeper.Point, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("robotgo location failed: %v", recovered)
		}
	}()
	point.X, point.Y = m.location()
	return point, nil
}

func (m *RobotMouse) MoveRelative(dx, dy int) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("robotgo move failed: %v", recovered)
		}
	}()
	m.move(dx, dy)
	return nil
}
