package app

import (
	"fannon-codder/internal/campaign"
	"fannon-codder/internal/data"
)

// Progress is the campaign loop shared by Boot Hill, briefing, and battle.
type Progress struct {
	Pool              *campaign.Pool
	Graves            int
	PhaseIndex        int
	MissionsCompleted int
	Campaign          *data.Campaign
	Phase             *data.Phase
	GameOver          bool
	AwaitingStub      bool // next mission has no map yet
	SaveNotice        string
	Heroes            []campaign.Hero
	// Field is the squad still out between phases of one mission.
	// Continue sends the debrief on to the next briefing, not Boot Hill.
	Field    []campaign.Soldier
	Continue bool
}

func NewProgress() *Progress {
	p := &Progress{Pool: campaign.NewGamePool()}
	c, err := data.LoadCampaign()
	if err == nil {
		p.Campaign = c
	}
	p.reloadPhase()
	return p
}

func (p *Progress) reloadPhase() {
	p.Phase = nil
	files := p.phaseFiles()
	if p.PhaseIndex < 0 || p.PhaseIndex >= len(files) {
		return
	}
	ph, err := data.LoadPhase(files[p.PhaseIndex])
	if err != nil {
		return
	}
	p.Phase = ph
}

func (p *Progress) DeployCount() int {
	if p.Phase != nil && p.Phase.Deploy > 0 {
		return p.Phase.Deploy
	}
	return 2
}

func (p *Progress) MissionNumber() int {
	if p.Phase != nil && p.Phase.Mission > 0 {
		return p.Phase.Mission
	}
	if p.MissionsCompleted > 0 {
		return p.MissionsCompleted
	}
	return 1
}

func (p *Progress) phaseFiles() []string {
	if p == nil || p.Campaign == nil {
		return nil
	}
	return p.Campaign.PhaseFiles()
}

func (p *Progress) MissionName() string {
	if p == nil || p.Campaign == nil {
		return ""
	}
	return p.Campaign.MissionName(p.MissionNumber())
}

func (p *Progress) available() int {
	n := len(p.Field)
	if p.Pool != nil {
		n += p.Pool.Remaining()
	}
	return n
}

func (p *Progress) CanStart() bool {
	if p.GameOver || p.AwaitingStub {
		return false
	}
	return p.Phase != nil && p.available() >= p.DeployCount()
}

func (p *Progress) CanSave() bool {
	return p.MissionsCompleted > 0 && !p.GameOver
}

func (p *Progress) MarkGameOverIfNeeded() {
	if p.AwaitingStub {
		return
	}
	if p.available() < p.DeployCount() {
		p.GameOver = true
	}
}

// AfterWin advances one phase. Survivors of a mission that still has
// parts stay out, topped up to the next deploy count. The last part
// sends them home and recruits.
func (p *Progress) AfterWin(survivors []campaign.Soldier) {
	if p == nil {
		return
	}
	cur := p.MissionNumber()
	p.OnPhaseWon()
	if p.Phase != nil && p.Phase.Mission == cur {
		p.fitField(survivors)
		return
	}
	p.Continue = false
	p.Field = nil
	if p.Pool != nil {
		p.Pool.Return(survivors)
	}
}

// fitField keeps the men who just lived, then takes or returns men so
// the squad matches this phase. Too few men is game over.
func (p *Progress) fitField(survivors []campaign.Soldier) {
	need := p.DeployCount()
	field := append([]campaign.Soldier{}, survivors...)
	if len(field) > need {
		p.Pool.Return(field[need:])
		field = append([]campaign.Soldier{}, field[:need]...)
	}
	if len(field) < need && p.Pool != nil {
		field = append(field, p.Pool.Deploy(need-len(field))...)
	}
	if len(field) < need {
		if p.Pool != nil {
			p.Pool.Return(field)
		}
		p.Field = nil
		p.Continue = false
		p.GameOver = true
		return
	}
	p.Field = field
	p.Continue = true
}

func (p *Progress) OnPhaseWon() {
	cur := p.MissionNumber()
	p.PhaseIndex++
	p.reloadPhase()
	if p.Phase != nil && p.Phase.Mission == cur {
		return // more phases in this mission
	}
	p.MissionsCompleted++
	p.Pool.CompleteMission(p.MissionsCompleted)
	if p.Phase == nil {
		p.AwaitingStub = true
	}
}

func (p *Progress) ToSave() campaign.SaveGame {
	var recruits []campaign.Soldier
	next := 0
	if p.Pool != nil {
		recruits = append(recruits, p.Pool.Recruits...)
		next = p.Pool.NextNameIndex()
	}
	return campaign.SaveGame{
		PhaseIndex:        p.PhaseIndex,
		MissionsCompleted: p.MissionsCompleted,
		Graves:            p.Graves,
		GameOver:          p.GameOver,
		AwaitingStub:      p.AwaitingStub,
		Recruits:          recruits,
		NextName:          next,
		Heroes:            append([]campaign.Hero{}, p.Heroes...),
	}
}

// NoteHero records a man who died with a personal score.
func (p *Progress) NoteHero(s campaign.Soldier) {
	if p == nil {
		return
	}
	p.Heroes = campaign.Remember(p.Heroes, campaign.Hero{Name: s.Name, Rank: s.Rank, Kills: s.Kills})
}

func (p *Progress) ApplySave(s campaign.SaveGame) {
	p.PhaseIndex = s.PhaseIndex
	p.MissionsCompleted = s.MissionsCompleted
	p.Graves = s.Graves
	p.GameOver = s.GameOver
	p.AwaitingStub = s.AwaitingStub
	p.Pool = campaign.RestorePool(s.Recruits, s.NextName)
	p.Heroes = append([]campaign.Hero{}, s.Heroes...)
	p.Field = nil
	p.Continue = false
	p.reloadPhase()
	// A save from before Mission 2 existed flagged the stub even though
	// phase index 1 is now a real map. Trust the loaded phase.
	if p.Phase != nil {
		p.AwaitingStub = false
	} else if p.MissionsCompleted > 0 {
		p.AwaitingStub = true
	}
}

func (p *Progress) Save() error {
	path, err := campaign.DefaultSavePath()
	if err != nil {
		return err
	}
	return campaign.Save(path, p.ToSave())
}

func (p *Progress) Load() error {
	path, err := campaign.DefaultSavePath()
	if err != nil {
		return err
	}
	s, err := campaign.Load(path)
	if err != nil {
		return err
	}
	p.ApplySave(s)
	return nil
}
