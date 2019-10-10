package core

import (
	"context"
	"github.com/clearmatics/autonity/log"
	"math/big"
	"sync"
	"time"

	"github.com/clearmatics/autonity/common"
)

const (
	initialProposeTimeout   = 3000 * time.Millisecond
	proposeTimeoutDelta     = 500 * time.Millisecond
	initialPrevoteTimeout   = 1000 * time.Millisecond
	prevoteTimeoutDelta     = 500 * time.Millisecond
	initialPrecommitTimeout = 1000 * time.Millisecond
	precommitTimeoutDelta   = 500 * time.Millisecond
)

type TimeoutEvent struct {
	roundWhenCalled  int64
	heightWhenCalled int64
	// message type: msgProposal msgPrevote	msgPrecommit
	step uint64
}

type timeout struct {
	timer   *time.Timer
	started bool
	step    Step
	sync.Mutex
}

func newTimeout(s Step) *timeout {
	return &timeout{
		started: false,
		step:    s,
	}
}

// runAfterTimeout() will be run in a separate go routine, so values used inside the function needs to be managed separately
func (t *timeout) scheduleTimeout(stepTimeout time.Duration, round int64, height int64, runAfterTimeout func(r int64, h int64)) {
	t.Lock()
	defer t.Unlock()
	t.started = true
	t.timer = time.AfterFunc(stepTimeout, func() {
		runAfterTimeout(round, height)
	})
}

func (t *timeout) timerStarted() bool {
	t.Lock()
	defer t.Unlock()
	return t.started
}

func (t *timeout) stopTimer() error {
	t.Lock()
	defer t.Unlock()
	if t.started {
		if t.started = !t.timer.Stop(); t.started {
			switch t.step {
			case propose:
				return errNilPrevoteSent
			case prevote:
				return errNilPrecommitSent
			case precommit:
				return errMovedToNewRound
			}
		}
	}
	return nil
}

func (t *timeout) reset(s Step) {
	err := t.stopTimer()
	if err != nil {
		log.Info("cant stop timer", "err", err)
	}

	t.Lock()
	defer t.Unlock()
	t.timer = nil
	t.started = false
	t.step = s
}

/////////////// On Timeout Functions ///////////////
func (c *core) onTimeoutPropose(r int64, h int64) {
	msg := TimeoutEvent{
		roundWhenCalled:  r,
		heightWhenCalled: h,
		step:             msgProposal,
	}
	c.logTimeoutEvent("TimeoutEvent(Propose): Sent", "Propose", msg)
	c.sendEvent(msg)
}

func (c *core) onTimeoutPrevote(r int64, h int64) {
	msg := TimeoutEvent{
		roundWhenCalled:  r,
		heightWhenCalled: h,
		step:             msgPrevote,
	}
	c.logTimeoutEvent("TimeoutEvent(Prevote): Sent", "Prevote", msg)
	c.sendEvent(msg)

}

func (c *core) onTimeoutPrecommit(r int64, h int64) {
	msg := TimeoutEvent{
		roundWhenCalled:  r,
		heightWhenCalled: h,
		step:             msgPrecommit,
	}
	c.logTimeoutEvent("TimeoutEvent(Precommit): Sent", "Precommit", msg)
	c.sendEvent(msg)
}

/////////////// Handle Timeout Functions ///////////////
func (c *core) handleTimeoutPropose(ctx context.Context, msg TimeoutEvent) {
	if msg.heightWhenCalled == c.currentRoundState.Height().Int64() && msg.roundWhenCalled == c.currentRoundState.Round().Int64() && c.currentRoundState.Step() == propose {
		c.logTimeoutEvent("TimeoutEvent(Propose): Received", "Propose", msg)
		c.sendPrevote(ctx, true)
		c.setStep(prevote)
	} else {
		c.logger.Error("got old timeout propose message",
			"Propose", msg,
			"msgHeight", msg.heightWhenCalled,
			"currentHeight", c.currentRoundState.Height().Int64(),
			"msgRound", msg.roundWhenCalled,
			"currentRound", c.currentRoundState.Round().Int64(),
			"currentRoundStep", c.currentRoundState.Step(),
			)
	}
}

func (c *core) handleTimeoutPrevote(ctx context.Context, msg TimeoutEvent) {
	if msg.heightWhenCalled == c.currentRoundState.Height().Int64() && msg.roundWhenCalled == c.currentRoundState.Round().Int64() && c.currentRoundState.Step() == prevote {
		c.logTimeoutEvent("TimeoutEvent(Prevote): Received", "Prevote", msg)
		c.sendPrecommit(ctx, true)
		c.setStep(precommit)
	} else {
		c.logger.Error("got old timeout prevote message",
			"Prevote", msg,
			"msgHeight", msg.heightWhenCalled,
			"currentHeight", c.currentRoundState.Height().Int64(),
			"msgRound", msg.roundWhenCalled,
			"currentRound", c.currentRoundState.Round().Int64(),
			"currentRoundStep", c.currentRoundState.Step(),
		)
	}
}

func (c *core) handleTimeoutPrecommit(ctx context.Context, msg TimeoutEvent) {
	if msg.heightWhenCalled == c.currentRoundState.Height().Int64() && msg.roundWhenCalled == c.currentRoundState.Round().Int64() {
		c.logTimeoutEvent("TimeoutEvent(Precommit): Received", "Precommit", msg)

		c.startRound(ctx, new(big.Int).Add(c.currentRoundState.Round(), common.Big1))
	} else {
		c.logger.Error("got old timeout preCommit message",
			"PreCommit", msg,
			"msgHeight", msg.heightWhenCalled,
			"currentHeight", c.currentRoundState.Height().Int64(),
			"msgRound", msg.roundWhenCalled,
			"currentRound", c.currentRoundState.Round().Int64(),
			"currentRoundStep", c.currentRoundState.Step(),
		)
	}
}

/////////////// Calculate Timeout Duration Functions ///////////////
// The timeout may need to be changed depending on the Step
func timeoutPropose(round int64) time.Duration {
	return initialProposeTimeout + time.Duration(round)*proposeTimeoutDelta
}

func timeoutPrevote(round int64) time.Duration {
	return initialPrevoteTimeout + time.Duration(round)*prevoteTimeoutDelta
}

func timeoutPrecommit(round int64) time.Duration {
	return initialPrecommitTimeout + time.Duration(round)*precommitTimeoutDelta
}

func (c *core) logTimeoutEvent(message string, msgType string, timeout TimeoutEvent) {
	c.logger.Debug(message,
		"from", c.address.String(),
		"type", msgType,
		"currentHeight", c.currentRoundState.Height(),
		"msgHeight", timeout.heightWhenCalled,
		"currentRound", c.currentRoundState.Round(),
		"msgRound", timeout.roundWhenCalled,
		"currentStep", c.currentRoundState.Step(),
		"msgStep", timeout.step,
	)
}
