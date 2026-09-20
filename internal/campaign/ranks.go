package campaign

// Rank is a trooper's promotion level. 0 is Private.
type Rank int

const (
	Private Rank = iota
	Corporal
	Sergeant
	StaffSergeant
	SergeantFirstClass
	MasterSergeant
	SergeantMajor
	Specialist4
	Specialist6
	WarrantOfficer
	ChiefWarrantOfficer
	Captain
	Major
	Colonel
	BrigadierGeneral
	General
	rankCount
)

var rankNames = [...]string{
	Private:             "Private",
	Corporal:            "Corporal",
	Sergeant:            "Sergeant",
	StaffSergeant:       "Staff Sergeant",
	SergeantFirstClass:  "Sergeant First Class",
	MasterSergeant:      "Master Sergeant",
	SergeantMajor:       "Sergeant Major",
	Specialist4:         "Specialist 4",
	Specialist6:         "Specialist 6",
	WarrantOfficer:      "Warrant Officer",
	ChiefWarrantOfficer: "Chief Warrant Officer",
	Captain:             "Captain",
	Major:               "Major",
	Colonel:             "Colonel",
	BrigadierGeneral:    "Brigadier General",
	General:             "General",
}

var rankAbbrev = [...]string{
	Private:             "Pte",
	Corporal:            "Cpl",
	Sergeant:            "Sgt",
	StaffSergeant:       "SSgt",
	SergeantFirstClass:  "SFC",
	MasterSergeant:      "MSG",
	SergeantMajor:       "SGM",
	Specialist4:         "SP4",
	Specialist6:         "SP6",
	WarrantOfficer:      "WO",
	ChiefWarrantOfficer: "CWO",
	Captain:             "Cpt",
	Major:               "Maj",
	Colonel:             "Col",
	BrigadierGeneral:    "BG",
	General:             "Gen",
}

func (r Rank) clamp() Rank {
	if r < 0 {
		return Private
	}
	if r >= rankCount {
		return General
	}
	return r
}

func (r Rank) String() string { return rankNames[r.clamp()] }

// Abbrev is the short HUD label (Pte, Cpl, …).
func (r Rank) Abbrev() string { return rankAbbrev[r.clamp()] }
