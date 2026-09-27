package sim

const (
	// ShallowSpeedMul: wading. Can still fire.
	ShallowSpeedMul = 0.5
	// DeepSpeedMul: swimming. Cannot fire (or throw).
	DeepSpeedMul = 1.0 / 3.0
)

func (w *World) speedMul(u *Unit) float64 {
	switch w.Map.TileAtPixel(u.X, u.Y) {
	case TileWaterShallow:
		return ShallowSpeedMul
	case TileWaterDeep:
		return DeepSpeedMul
	default:
		return 1
	}
}

// CanShoot is false in deep water (Amiga: swimming troopers cannot fire)
// and while sinking in quicksand.
func (w *World) CanShoot(u *Unit) bool {
	if u == nil || !u.Living() || u.Sinking || u.VehicleID != 0 {
		return false
	}
	return w.Map.TileAtPixel(u.X, u.Y) != TileWaterDeep
}

func (w *World) refreshTerrain() {
	for i := range w.Units {
		u := &w.Units[i]
		t := w.Map.TileAtPixel(u.X, u.Y)
		wet := t == TileWaterShallow || t == TileWaterDeep
		// Crew ride the hull. Wading and swimming are the same splash.
		if wet && !u.InWater && u.sampled && u.VehicleID == 0 {
			w.emit(CueSplash, u.X, u.Y)
		}
		u.InWater = wet
		u.sampled = true
	}
}
