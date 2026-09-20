package app

import (
	"fannon-codder/internal/campaign"
	"fannon-codder/internal/data"
)

// Progress is the campaign loop shared by Boot Hill, briefing, and battle.
type Progress struct {
	Pool         *campaign.Pool
	Graves       int
	PhaseIndex   int
	Campaign     *data.Campaign
	Phase        *data.Phase
	MissionDone  bool
	GameOver     bool
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
	return 1
}

func (p *Progress) CanStart() bool {
	return !p.MissionDone && !p.GameOver && p.Pool.Remaining() >= p.DeployCount()
}

func (p *Progress) MarkGameOverIfNeeded() {
	if p.Pool.Remaining() < p.DeployCount() {
		p.GameOver = true
	}
}

func (p *Progress) OnPhaseWon() {
	if p.Campaign == nil {
		p.MissionDone = true
		return
	}
	cur := p.MissionNumber()
	p.PhaseIndex++
	p.reloadPhase()
	if p.Phase == nil || p.Phase.Mission != cur {
		// Next file is a new mission (or none). Promotions/+15 are chunk 09.
		p.MissionDone = true
		p.PhaseIndex-- // stay on completed phase for the label
		p.reloadPhase()
	}
}
