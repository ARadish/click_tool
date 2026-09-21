package desktop

import "testing"

func TestRobotMouseConvertsLocationToPoint(t *testing.T) {
	mouse := newRobotMouse(func() (int, int) { return 12, 34 }, func(int, int) {})
	point, err := mouse.Position()
	if err != nil {
		t.Fatalf("Position returned error: %v", err)
	}
	if point.X != 12 || point.Y != 34 {
		t.Fatalf("Position = %#v, want X=12 Y=34", point)
	}
}

func TestRobotMouseReportsAutomationPanicAsError(t *testing.T) {
	mouse := newRobotMouse(func() (int, int) { panic("blocked") }, func(int, int) {})
	if _, err := mouse.Position(); err == nil {
		t.Fatal("Position succeeded, want recovered automation error")
	}

	mouse = newRobotMouse(func() (int, int) { return 0, 0 }, func(int, int) { panic("blocked") })
	if err := mouse.MoveRelative(1, 0); err == nil {
		t.Fatal("MoveRelative succeeded, want recovered automation error")
	}
}
