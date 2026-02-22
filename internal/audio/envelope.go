package audio

// ADSR represents an Attack-Decay-Sustain-Release envelope.
type ADSR struct {
	Attack  float64 // seconds
	Decay   float64 // seconds
	Sustain float64 // level 0.0-1.0
	Release float64 // seconds
}

// Level returns the envelope amplitude at the given time.
// noteOnDuration is how long the note is held before release begins.
func (e ADSR) Level(time, noteOnDuration float64) float64 {
	if time < 0 {
		return 0
	}

	if time < noteOnDuration {
		// Note-on phase: attack -> decay -> sustain.
		if time < e.Attack {
			// Attack: ramp 0 -> 1.
			if e.Attack == 0 {
				return 1
			}
			return time / e.Attack
		}
		time -= e.Attack

		if time < e.Decay {
			// Decay: ramp 1 -> sustain.
			if e.Decay == 0 {
				return e.Sustain
			}
			return 1 - (1-e.Sustain)*(time/e.Decay)
		}

		// Sustain.
		return e.Sustain
	}

	// Release phase.
	releaseTime := time - noteOnDuration
	if e.Release <= 0 || releaseTime >= e.Release {
		return 0
	}
	// Calculate the level at the moment of release.
	releaseLevel := e.levelAtRelease(noteOnDuration)
	return releaseLevel * (1 - releaseTime/e.Release)
}

// levelAtRelease returns the envelope level at the moment the note is released.
func (e ADSR) levelAtRelease(noteOnDuration float64) float64 {
	if noteOnDuration < e.Attack {
		if e.Attack == 0 {
			return 1
		}
		return noteOnDuration / e.Attack
	}
	t := noteOnDuration - e.Attack
	if t < e.Decay {
		if e.Decay == 0 {
			return e.Sustain
		}
		return 1 - (1-e.Sustain)*(t/e.Decay)
	}
	return e.Sustain
}
