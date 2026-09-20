package sim

// SquadID is Snake, Eagle, or Panther (max three player groups).
type SquadID int

const (
	SquadSnake SquadID = iota
	SquadEagle
	SquadPanther
)

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
