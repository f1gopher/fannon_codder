package sim

// SquadID is Snake, Eagle, or Panther (max three player groups).
type SquadID int

const (
	SquadSnake SquadID = iota
	SquadEagle
	SquadPanther
)

// Letter is the HUD mark: S, E, or P.
func (id SquadID) Letter() string {
	switch id {
	case SquadEagle:
		return "E"
	case SquadPanther:
		return "P"
	default:
		return "S"
	}
}

// AmmoShare is how many grenades or rockets a split hands the new squad.
// Cycled from the HUD icons; ammo stays 0 until grenades exist.
type AmmoShare int

const (
	ShareNone AmmoShare = iota
	ShareHalf
	ShareAll
)

// Next cycles none → half → all → none.
func (a AmmoShare) Next() AmmoShare {
	switch a {
	case ShareNone:
		return ShareHalf
	case ShareHalf:
		return ShareAll
	default:
		return ShareNone
	}
}

// MergeRadius is how close two troopers from different squads must be
// (centre to centre) before those squads combine. Under FileSpacing so a
// squad standing in file does not merge with itself after a split.
const MergeRadius = 8

// Squad is one ordered file of troopers. MemberIDs[0] is the leader.
type Squad struct {
	ID        SquadID
	LeaderID  int
	MemberIDs []int
	Grenades  int
	Rockets   int
	Active    bool

	DestX, DestY float64
	HasDest      bool
	Trail        []Vec2
}

// Vec2 is a world-space point (trail crumbs, destinations).
type Vec2 struct {
	X, Y float64
}
