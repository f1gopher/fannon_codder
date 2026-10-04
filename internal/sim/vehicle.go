package sim

import "math"

const (
	VehicleCapacity = 8
	VehicleMinSpeed = 28.0
	VehicleMaxSpeed = 88.0
	VehicleHoldFull = 1.4 // seconds of hold to reach top speed
	VehicleMGRange  = 110.0
	VehicleMGRoF    = 8.0
	VehicleBoardR   = 16.0
	VehicleRamR     = 12.0
	vehicleHalf     = 7.0
	iceSkidRate     = 0.55 // per second; lower keeps more speed on ice
)

// Hover is the pointer affordance over a skidoo.
type Hover int

const (
	HoverNone Hover = iota
	HoverBoard
	HoverExit
)

// Vehicle is a skidoo. Jeeps later reuse this with a different skin.
type Vehicle struct {
	ID        int
	X, Y      float64
	VX, VY    float64
	Facing    float64
	Side      Side
	Armed     bool
	Alive     bool
	Capacity  int
	Occupants []int
	Hold      float64
	FireCD    float64
	Blink     float64
}

// AddSkidoo parks a vehicle. An enemy skidoo starts with a driver aboard.
func (w *World) AddSkidoo(p Vec2, side Side, armed bool) *Vehicle {
	if w.nextVID == 0 {
		w.nextVID = 1
	}
	v := Vehicle{
		ID:       w.nextVID,
		X:        p.X,
		Y:        p.Y,
		Side:     side,
		Armed:    armed,
		Alive:    true,
		Capacity: VehicleCapacity,
	}
	w.nextVID++
	w.Vehicles = append(w.Vehicles, v)
	out := &w.Vehicles[len(w.Vehicles)-1]
	if side == SideEnemy {
		u := w.SpawnUnit(SideEnemy, p)
		u.VehicleID = out.ID
		out.Occupants = []int{u.ID}
	}
	return out
}

func (w *World) vehicleByID(id int) *Vehicle {
	if id == 0 {
		return nil
	}
	for i := range w.Vehicles {
		if w.Vehicles[i].ID == id {
			return &w.Vehicles[i]
		}
	}
	return nil
}

func (w *World) vehicleOccupied(v *Vehicle) bool {
	if v == nil {
		return false
	}
	for _, id := range v.Occupants {
		if u := w.Unit(id); u != nil && u.Living() {
			return true
		}
	}
	return false
}

func (w *World) LeaderInVehicle() bool {
	s := w.ActiveSquad()
	if s == nil {
		return false
	}
	u := w.Unit(s.LeaderID)
	return u != nil && u.Living() && u.VehicleID != 0
}

// VehicleHover reports board or exit when the pointer is over a skidoo.
func (w *World) VehicleHover(x, y float64) (Hover, *Vehicle) {
	s := w.ActiveSquad()
	var leader *Unit
	if s != nil {
		leader = w.Unit(s.LeaderID)
	}
	var best *Vehicle
	bestD := VehicleBoardR
	for i := range w.Vehicles {
		v := &w.Vehicles[i]
		if !v.Alive {
			continue
		}
		d := hypot(v.X-x, v.Y-y)
		if d > bestD {
			continue
		}
		bestD = d
		best = v
	}
	if best == nil {
		return HoverNone, nil
	}
	if leader != nil && leader.Living() && leader.VehicleID == best.ID {
		return HoverExit, best
	}
	if !w.vehicleOccupied(best) {
		return HoverBoard, best
	}
	return HoverNone, best
}

// SetDrive is held left while the active squad is mounted.
// holding false coasts; on ice the old velocity persists.
func (w *World) SetDrive(x, y float64, holding bool) {
	w.Driving = holding
	w.DriveX = x
	w.DriveY = y
}

// CommandBoard walks the active squad onto an empty skidoo, or enters now if close.
func (w *World) CommandBoard(id int) bool {
	v := w.vehicleByID(id)
	s := w.ActiveSquad()
	if v == nil || s == nil || !v.Alive || w.vehicleOccupied(v) {
		return false
	}
	leader := w.Unit(s.LeaderID)
	if leader == nil || !leader.Living() || leader.VehicleID != 0 {
		return false
	}
	w.CommandMove(v.X, v.Y, true)
	w.BoardID = v.ID
	if hypot(leader.X-v.X, leader.Y-v.Y) <= VehicleBoardR+8 {
		return w.enterVehicle(v)
	}
	return true
}

// CommandExit puts the mounted squad back on foot beside the skidoo.
func (w *World) CommandExit(id int) bool {
	v := w.vehicleByID(id)
	s := w.ActiveSquad()
	if v == nil || s == nil || !v.Alive {
		return false
	}
	leader := w.Unit(s.LeaderID)
	if leader == nil || leader.VehicleID != v.ID {
		return false
	}
	w.dismount(v)
	return true
}

func (w *World) tryCompleteBoard() {
	if w.BoardID == 0 {
		return
	}
	v := w.vehicleByID(w.BoardID)
	s := w.ActiveSquad()
	if v == nil || !v.Alive || s == nil || w.vehicleOccupied(v) {
		w.BoardID = 0
		return
	}
	leader := w.Unit(s.LeaderID)
	if leader == nil || !leader.Living() {
		w.BoardID = 0
		return
	}
	if hypot(leader.X-v.X, leader.Y-v.Y) <= VehicleBoardR+8 {
		w.enterVehicle(v)
	}
}

func (w *World) enterVehicle(v *Vehicle) bool {
	s := w.ActiveSquad()
	if s == nil || v == nil || !v.Alive || w.vehicleOccupied(v) {
		return false
	}
	capn := v.Capacity
	if capn <= 0 {
		capn = VehicleCapacity
	}
	var ids []int
	for _, id := range s.MemberIDs {
		u := w.Unit(id)
		if u == nil || !u.Living() || u.VehicleID != 0 {
			continue
		}
		ids = append(ids, id)
		if len(ids) >= capn {
			break
		}
	}
	if len(ids) == 0 {
		return false
	}
	for _, id := range ids {
		u := w.Unit(id)
		u.VehicleID = v.ID
		u.X, u.Y = v.X, v.Y
		u.VX, u.VY = 0, 0
		u.Sinking = false
	}
	v.Occupants = append(v.Occupants, ids...)
	v.Side = SidePlayer
	s.HasDest = false
	w.BoardID = 0
	w.emit(CueBoard, v.X, v.Y)
	return true
}

func (w *World) dismount(v *Vehicle) {
	n := 0
	for _, id := range v.Occupants {
		u := w.Unit(id)
		if u == nil || !u.Living() {
			continue
		}
		pos := w.exitSpot(v, n)
		u.X, u.Y = pos.X, pos.Y
		u.VehicleID = 0
		u.VX, u.VY = 0, 0
		n++
	}
	v.Occupants = nil
	v.Hold = 0
	v.VX, v.VY = 0, 0
	w.emit(CueExit, v.X, v.Y)
}

func (w *World) exitSpot(v *Vehicle, n int) Vec2 {
	spots := []Vec2{
		{0, 18}, {-12, 18}, {12, 18}, {-22, 18}, {22, 18}, {0, 28},
	}
	off := spots[n%len(spots)]
	x, y := v.X+off.X, v.Y+off.Y
	if w.walkableUnit(x, y) {
		return Vec2{X: x, Y: y}
	}
	return Vec2{X: v.X + off.X, Y: v.Y + off.Y}
}

func (w *World) destroyVehicle(v *Vehicle) {
	if v == nil || !v.Alive {
		return
	}
	v.Alive = false
	v.VX, v.VY = 0, 0
	for _, id := range v.Occupants {
		if u := w.Unit(id); u != nil {
			w.kill(u)
		}
	}
	v.Occupants = nil
}

func (w *World) stepVehicles(dt float64) {
	for i := range w.Vehicles {
		v := &w.Vehicles[i]
		if !v.Alive {
			continue
		}
		v.Blink += dt
		if v.Side == SideEnemy && w.AI {
			w.stepEnemyVehicle(v, dt)
		} else if w.leaderVehicle() == v {
			w.applyDrive(v, w.DriveX, w.DriveY, w.Driving, dt)
			if w.Firing && v.Armed {
				w.vehicleShoot(v, w.AimX, w.AimY, dt)
			}
		} else {
			w.applyDrive(v, v.X, v.Y, false, dt)
		}
		w.syncCrew(v)
		w.ram(v)
	}
	w.Driving = false
}

func (w *World) leaderVehicle() *Vehicle {
	s := w.ActiveSquad()
	if s == nil {
		return nil
	}
	u := w.Unit(s.LeaderID)
	if u == nil || u.VehicleID == 0 {
		return nil
	}
	return w.vehicleByID(u.VehicleID)
}

// stepEnemyVehicle drives at the nearest living player. An unarmed hull
// keeps going and can ram. An armed hull drops the throttle once it is
// inside VehicleMGRange, which is still outside the ram, and turns the
// hull toward that player. The mounted gun then uses the grunt reaction,
// facing tolerance, burst, and pause. Range and rate stay the vehicle's.
func (w *World) stepEnemyVehicle(v *Vehicle, dt float64) {
	px, py, ok := w.nearestLiving(SidePlayer, v.X, v.Y)
	if !ok {
		w.applyDrive(v, v.X, v.Y, false, dt)
		w.idleVehicleGun(v)
		return
	}
	if !v.Armed {
		w.applyDrive(v, px, py, true, dt)
		return
	}
	dist := hypot(px-v.X, py-v.Y)
	inRange := dist <= VehicleMGRange
	w.applyDrive(v, px, py, !inRange, dt)
	if inRange {
		turnVehicle(v, px, py, dt)
	}
	w.stepVehicleGun(v, px, py, inRange && w.lineClear(v.X, v.Y, px, py), dt)
}

// turnVehicle swings the hull at the grunt turn rate. Drive code owns
// facing while the throttle is held; this runs on the gun line.
func turnVehicle(v *Vehicle, px, py, dt float64) {
	dx, dy := px-v.X, py-v.Y
	if dx == 0 && dy == 0 {
		return
	}
	d := wrapAngle(math.Atan2(dy, dx) - v.Facing)
	step := EnemyTurnRate * dt
	if d > step {
		d = step
	} else if d < -step {
		d = -step
	}
	v.Facing = wrapAngle(v.Facing + d)
}

func vehicleFacingError(v *Vehicle, px, py float64) float64 {
	dx, dy := px-v.X, py-v.Y
	if dx == 0 && dy == 0 {
		return 0
	}
	return math.Abs(wrapAngle(math.Atan2(dy, dx) - v.Facing))
}

func (w *World) vehicleGunner(v *Vehicle) *Unit {
	if v == nil {
		return nil
	}
	for _, id := range v.Occupants {
		if u := w.Unit(id); u != nil && u.Living() {
			return u
		}
	}
	return nil
}

func (w *World) idleVehicleGun(v *Vehicle) {
	if u := w.vehicleGunner(v); u != nil {
		resetGruntContact(u)
	}
	v.FireCD = 0
}

// stepVehicleGun is the mounted burst. The clock lives on the driver.
// Losing range or LOS clears it. The round itself is still vehicleShoot's
// gun: VehicleMGRange, VehicleMGRoF, no grunt cone.
func (w *World) stepVehicleGun(v *Vehicle, px, py float64, see bool, dt float64) {
	u := w.vehicleGunner(v)
	if u == nil || !see {
		w.idleVehicleGun(v)
		return
	}
	if u.ReactAt == 0 {
		if w.rng == nil {
			w.rng = newRNG()
		}
		u.ReactAt = EnemyReact + (w.rng.Float64()*2-1)*EnemyReactJitter
	}
	u.SpotT += dt
	if u.SpotT+1e-9 < u.ReactAt {
		return
	}
	if u.BurstGap > 0 {
		u.BurstGap -= dt
		if u.BurstGap > 0 {
			return
		}
		u.BurstGap = 0
		u.BurstN = 0
		v.FireCD = 0
	}
	if vehicleFacingError(v, px, py) > EnemyFaceTol {
		return
	}
	v.FireCD -= dt
	if v.FireCD > 0 {
		return
	}
	u.BurstN++
	if u.BurstN >= EnemyBurst {
		u.BurstGap = EnemyBurstPause
		v.FireCD = 0
	} else {
		v.FireCD = 1.0 / VehicleMGRoF
	}
	w.fireVehicleMG(v, v.Facing)
}

func (w *World) applyDrive(v *Vehicle, tx, ty float64, holding bool, dt float64) {
	ice := w.Map.TileAtPixel(v.X, v.Y) == TileIce
	if holding {
		v.Hold += dt
		t := v.Hold / VehicleHoldFull
		if t > 1 {
			t = 1
		}
		speed := VehicleMinSpeed + (VehicleMaxSpeed-VehicleMinSpeed)*t
		if w.Map.TileAtPixel(v.X, v.Y) == TileWaterShallow {
			speed *= ShallowSpeedMul
		}
		dx, dy := tx-v.X, ty-v.Y
		dist := hypot(dx, dy)
		if dist < 1 {
			dx, dy = 1, 0
			dist = 1
		}
		desX := dx / dist * speed
		desY := dy / dist * speed
		blend := 1.0
		if ice {
			blend = 3 * dt
			if blend > 1 {
				blend = 1
			}
		}
		v.VX += (desX - v.VX) * blend
		v.VY += (desY - v.VY) * blend
		v.Facing = math.Atan2(v.VY, v.VX)
	} else {
		v.Hold = 0
		if ice {
			decay := math.Exp(-iceSkidRate * dt)
			v.VX *= decay
			v.VY *= decay
			if hypot(v.VX, v.VY) < 6 {
				v.VX, v.VY = 0, 0
			}
		} else {
			v.VX, v.VY = 0, 0
		}
	}
	w.slideVehicle(v, v.VX*dt, v.VY*dt)
}

func (w *World) slideVehicle(v *Vehicle, dx, dy float64) {
	if w.vehicleWalkable(v.X+dx, v.Y+dy) {
		v.X += dx
		v.Y += dy
		return
	}
	if w.vehicleWalkable(v.X+dx, v.Y) {
		v.X += dx
		v.VY = 0
		return
	}
	if w.vehicleWalkable(v.X, v.Y+dy) {
		v.Y += dy
		v.VX = 0
		return
	}
	v.VX, v.VY = 0, 0
}

func (w *World) vehicleWalkable(x, y float64) bool {
	if !w.Map.Walkable(x, y, vehicleHalf) {
		return false
	}
	if w.Map.TileAtPixel(x, y) == TileWaterDeep {
		return false
	}
	return !w.unitHitsBuilding(x, y, vehicleHalf)
}

func (w *World) syncCrew(v *Vehicle) {
	for _, id := range v.Occupants {
		u := w.Unit(id)
		if u == nil || !u.Living() {
			continue
		}
		u.X, u.Y = v.X, v.Y
		u.VX, u.VY = 0, 0
	}
}

func (w *World) vehicleDriver(v *Vehicle) int {
	if v == nil {
		return 0
	}
	for _, id := range v.Occupants {
		if u := w.Unit(id); u != nil && u.Living() {
			return u.ID
		}
	}
	return 0
}

func (w *World) ram(v *Vehicle) {
	if !v.Alive || hypot(v.VX, v.VY) < 8 {
		return
	}
	for i := range w.Units {
		u := &w.Units[i]
		if !u.Living() || u.VehicleID != 0 {
			continue
		}
		if hypot(u.X-v.X, u.Y-v.Y) <= VehicleRamR {
			w.kill(u)
			w.creditKill(w.vehicleDriver(v), u)
		}
	}
}

func (w *World) vehicleShoot(v *Vehicle, tx, ty, dt float64) {
	if v == nil || !v.Alive || !v.Armed {
		return
	}
	v.FireCD -= dt
	if v.FireCD > 0 {
		return
	}
	v.FireCD = 1.0 / VehicleMGRoF
	dx, dy := tx-v.X, ty-v.Y
	if dx == 0 && dy == 0 {
		dx = 1
	}
	ang := math.Atan2(dy, dx)
	v.Facing = ang
	w.fireVehicleMG(v, ang)
}

func (w *World) fireVehicleMG(v *Vehicle, ang float64) {
	owner := w.vehicleDriver(v)
	side := v.Side
	if u := w.Unit(owner); u != nil {
		side = u.Side
	}
	w.addMG(v.X+math.Cos(ang)*10, v.Y+math.Sin(ang)*10, ang, VehicleMGRange, owner, side)
}
