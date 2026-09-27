package app

import (
	"testing"

	"fannon-codder/internal/campaign"
)

func TestCleanMission1ThenMission2Shows27(t *testing.T) {
	p := NewProgress()
	if p.DeployCount() != 2 || p.MissionNumber() != 1 {
		t.Fatalf("start deploy=%d mission=%d", p.DeployCount(), p.MissionNumber())
	}
	men := p.Pool.Deploy(p.DeployCount())
	if p.Pool.Remaining() != 13 {
		t.Fatalf("M1 remaining=%d, want 13", p.Pool.Remaining())
	}
	for i := range men {
		men[i].PhasesThisMission++
	}
	p.Pool.Return(men)
	p.OnPhaseWon()
	if p.MissionsCompleted != 1 || p.AwaitingStub {
		t.Fatalf("after M1: completed=%d stub=%v", p.MissionsCompleted, p.AwaitingStub)
	}
	if p.Phase == nil || p.Phase.Mission != 2 || p.Phase.Phase != 1 {
		t.Fatal("Boot Hill should offer Mission 2 phase 1")
	}
	if p.Pool.Remaining() != 30 {
		t.Fatalf("pool after clean M1=%d, want 30", p.Pool.Remaining())
	}
	m2 := p.Pool.Deploy(p.DeployCount())
	if len(m2) != 3 || p.Pool.Remaining() != 27 {
		t.Fatalf("M2 deploy %d, remaining %d; want 3 and 27", len(m2), p.Pool.Remaining())
	}
	cpls := 0
	for _, s := range m2 {
		if s.Rank == campaign.Corporal {
			cpls++
		}
	}
	if cpls != 2 {
		t.Fatalf("M1 survivors should deploy first as corporals, got %d", cpls)
	}

	// Both Mission 2 phases, no deaths: +2 ranks, then Mission 3.
	for i := range m2 {
		m2[i].PhasesThisMission++
	}
	p.Pool.Return(m2)
	p.OnPhaseWon()
	if p.MissionsCompleted != 1 || p.Phase == nil || p.Phase.Phase != 2 {
		t.Fatalf("phase 2 not armed: completed=%d phase=%v", p.MissionsCompleted, p.Phase)
	}
	m2b := p.Pool.Deploy(p.DeployCount())
	for i := range m2b {
		m2b[i].PhasesThisMission++
	}
	p.Pool.Return(m2b)
	p.OnPhaseWon()
	if p.MissionsCompleted != 2 || p.AwaitingStub {
		t.Fatalf("after M2: completed=%d stub=%v", p.MissionsCompleted, p.AwaitingStub)
	}
	if p.Phase == nil || p.Phase.Mission != 3 || p.DeployCount() != 4 {
		t.Fatalf("Mission 3 not armed: %+v deploy=%d", p.Phase, p.DeployCount())
	}
	if p.Pool.Remaining() != 45 {
		t.Fatalf("pool after clean M2=%d, want 45", p.Pool.Remaining())
	}
	staff, sgts := 0, 0
	for _, s := range p.Pool.Recruits {
		switch s.Rank {
		case campaign.StaffSergeant:
			staff++
		case campaign.Sergeant:
			sgts++
		}
	}
	if staff != 2 || sgts < 1 {
		t.Fatalf("full M2 survivors: staff=%d sgt=%d, want 2 and at least 1", staff, sgts)
	}
	q := NewProgress()
	q.ApplySave(p.ToSave())
	if q.AwaitingStub || q.Phase == nil || q.Phase.Mission != 3 {
		t.Fatal("a finished Mission 2 save should open Mission 3")
	}
	m3 := p.Pool.Deploy(p.DeployCount())
	for i := range m3 {
		m3[i].PhasesThisMission++
	}
	p.Pool.Return(m3)
	p.OnPhaseWon()
	if p.MissionsCompleted != 3 || p.AwaitingStub {
		t.Fatalf("after M3: completed=%d stub=%v", p.MissionsCompleted, p.AwaitingStub)
	}
	if p.Phase == nil || p.Phase.Mission != 4 || p.Phase.Phase != 1 || p.DeployCount() != 4 {
		t.Fatalf("Mission 4 not armed: %+v deploy=%d", p.Phase, p.DeployCount())
	}
	q4 := NewProgress()
	q4.ApplySave(p.ToSave())
	if q4.AwaitingStub || q4.Phase == nil || q4.Phase.Title != "Beachy Head" {
		t.Fatal("a finished Mission 3 save should open Beachy Head")
	}
	// Four Mission 4 phases, then the Mission 5 stub.
	for i, want := range []struct {
		phase, deploy, grenades int
	}{
		{1, 4, 0},
		{2, 4, 2},
		{3, 5, 2},
		{4, 5, 2},
	} {
		if p.Phase == nil || p.Phase.Phase != want.phase || p.DeployCount() != want.deploy {
			t.Fatalf("M4 phase index %d: %+v deploy=%d", i, p.Phase, p.DeployCount())
		}
		if p.Phase.StartGrenadesPerTrooper != want.grenades {
			t.Fatalf("M4 phase %d start grenades=%d", want.phase, p.Phase.StartGrenadesPerTrooper)
		}
		men := p.Pool.Deploy(p.DeployCount())
		for j := range men {
			men[j].PhasesThisMission++
		}
		p.Pool.Return(men)
		p.OnPhaseWon()
	}
	if p.MissionsCompleted != 4 || !p.AwaitingStub || p.Phase != nil {
		t.Fatalf("after M4: completed=%d stub=%v phase=%v", p.MissionsCompleted, p.AwaitingStub, p.Phase)
	}
}

func TestOldStubSaveLoadsMission2(t *testing.T) {
	p := NewProgress()
	men := p.Pool.Deploy(2)
	p.Pool.Return(men)
	p.OnPhaseWon()
	save := p.ToSave()
	save.AwaitingStub = true // chunk 09 saves, from before this map existed
	q := NewProgress()
	q.ApplySave(save)
	if q.AwaitingStub || q.Phase == nil || q.Phase.Mission != 2 {
		t.Fatalf("stub=%v phase=%v", q.AwaitingStub, q.Phase)
	}
}
