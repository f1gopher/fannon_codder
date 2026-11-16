package app

import (
	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/input"
)

// padPanSpeed is how fast the right stick scrolls the battle, in world pixels per second.
const padPanSpeed = 220

// readPad samples the first connected standard-layout gamepad.
func readPad() input.Pad {
	for _, id := range ebiten.AppendGamepadIDs(nil) {
		if !ebiten.IsStandardGamepadLayoutAvailable(id) {
			continue
		}
		return input.Pad{
			Connected: true,
			StickX:    axis(id, ebiten.StandardGamepadAxisLeftStickHorizontal),
			StickY:    axis(id, ebiten.StandardGamepadAxisLeftStickVertical),
			RightX:    axis(id, ebiten.StandardGamepadAxisRightStickHorizontal),
			RightY:    axis(id, ebiten.StandardGamepadAxisRightStickVertical),
			South:     down(id, ebiten.StandardGamepadButtonRightBottom),
			East:      down(id, ebiten.StandardGamepadButtonRightRight),
			West:      down(id, ebiten.StandardGamepadButtonRightLeft),
			North:     down(id, ebiten.StandardGamepadButtonRightTop),
			LB:        down(id, ebiten.StandardGamepadButtonFrontTopLeft),
			RB:        down(id, ebiten.StandardGamepadButtonFrontTopRight),
			LT:        down(id, ebiten.StandardGamepadButtonFrontBottomLeft),
			RT:        down(id, ebiten.StandardGamepadButtonFrontBottomRight),
			Start:     down(id, ebiten.StandardGamepadButtonCenterRight),
			Back:      down(id, ebiten.StandardGamepadButtonCenterLeft),
			Up:        down(id, ebiten.StandardGamepadButtonLeftTop),
			Down:      down(id, ebiten.StandardGamepadButtonLeftBottom),
			Left:      down(id, ebiten.StandardGamepadButtonLeftLeft),
			Right:     down(id, ebiten.StandardGamepadButtonLeftRight),
		}
	}
	return input.Pad{}
}

func axis(id ebiten.GamepadID, a ebiten.StandardGamepadAxis) float64 {
	return ebiten.StandardGamepadAxisValue(id, a)
}

func down(id ebiten.GamepadID, b ebiten.StandardGamepadButton) bool {
	return ebiten.IsStandardGamepadButtonPressed(id, b)
}
