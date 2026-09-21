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
	if p.Campaign == nil || p.PhaseIndex < 0 || p.PhaseIndex >= len(p.Campaign.Phases) {
		return
	}
	ph, err := data.LoadPhase(p.Campaign.Phases[p.PhaseIndex])
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

func (p *Progress) CanStart() bool {
	if p.GameOver || p.AwaitingStub {
		return false
	}
	return p.Phase != nil && p.Pool.Remaining() >= p.DeployCount()
}

func (p *Progress) CanSave() bool {
	return p.MissionsCompleted > 0 && !p.GameOver
}

func (p *Progress) MarkGameOverIfNeeded() {
	if p.AwaitingStub {
		return
	}
	if p.Pool.Remaining() < p.DeployCount() {
		p.GameOver = true
	}
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
	}
}

func (p *Progress) ApplySave(s campaign.SaveGame) {
	p.PhaseIndex = s.PhaseIndex
	p.MissionsCompleted = s.MissionsCompleted
	p.Graves = s.Graves
	p.GameOver = s.GameOver
	p.AwaitingStub = s.AwaitingStub
	p.Pool = campaign.RestorePool(s.Recruits, s.NextName)
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
