package gui

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
)

// simulationLoop runs the main simulation loop at ~60 FPS
func (a *App) simulationLoop() {
	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	lastTime := time.Now()

	for a.running {
		<-ticker.C

		if a.paused {
			continue
		}

		if a.material == nil {
			continue
		}

		dt := time.Since(lastTime).Seconds()
		lastTime = time.Now()
		a.simTime += dt
		// Wrap simTime to prevent floating-point issues after long runs
		if a.simTime > 1000 {
			a.simTime = math.Mod(a.simTime, 1000)
		}

		a.mu.Lock()

		// Generate E-field based on waveform
		if a.waveform == WaveformManual {
			// Manual mode: slider control or click-to-level animation
			//
			// PHYSICS: Hysteresis is PATH-DEPENDENT and NON-REVERSIBLE.
			// If you overshoot a target level, you CANNOT correct by applying less field
			// or opposite field (that's a different branch of the hysteresis loop).
			// You MUST reset to a known saturation state and try again.
			//
			// Phases:
			// 0: RESET - saturate in opposite direction to target
			// 1: HOLD_RESET - return to zero (now at known remanent: level 1 or 30)
			// 2: WRITE - apply calibrated field toward target
			// 3: HOLD_WRITE - return to zero, polarization persists at target
			if a.manualAnimating {
				Ec := a.material.Ec
				Emax := Ec * 2.0
				phaseDuration := 0.6 / a.frequency
				rampRate := 4.0 * Emax * a.frequency

				a.manualPhaseTime += dt

				targetLevel := a.manualTargetLevel // 1-indexed (1-30)
				startLevel := a.manualStartLevel   // Captured at animation start

				switch a.manualPhase {
				case 0: // RESET - saturate in opposite direction to target
					var resetE float64
					if targetLevel > startLevel {
						// Going UP: first saturate negative (reach level 1)
						resetE = -2.0 * Ec
					} else {
						// Going DOWN: first saturate positive (reach level 30)
						resetE = 2.0 * Ec
					}

					// Ramp to reset field
					diff := resetE - a.electricField
					step := rampRate * dt
					if math.Abs(diff) < step {
						a.electricField = resetE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}

					// Transition when field reached and held briefly
					if a.manualPhaseTime > phaseDuration*0.3 && math.Abs(a.electricField-resetE) < 0.01*Emax {
						a.manualPhase = 1
						a.manualPhaseTime = 0
					}

				case 1: // HOLD_RESET - return to zero (now at known remanent state)
					step := rampRate * dt
					if math.Abs(a.electricField) < step {
						a.electricField = 0
					} else if a.electricField > 0 {
						a.electricField -= step
					} else {
						a.electricField += step
					}

					// Now at known remanent state (level 1 or level 30)
					if a.manualPhaseTime > phaseDuration*0.2 && math.Abs(a.electricField) < 0.01*Emax {
						a.manualPhase = 2
						a.manualPhaseTime = 0
					}

				case 2: // WRITE - apply calibrated field for target
					var writeE float64
					if targetLevel > startLevel {
						// Going UP from level 1: use ascending calibration
						writeE = a.calibrationUp[targetLevel-1] // 0-indexed array
						if writeE == 0 {
							// Fallback: interpolate based on target position
							ratio := float64(targetLevel-1) / 29.0
							writeE = Ec * (1.0 + ratio*1.0) // Ec to 2*Ec
						}
					} else {
						// Going DOWN from level 30: use descending calibration
						writeE = a.calibrationDown[targetLevel-1] // 0-indexed array
						if writeE == 0 {
							ratio := float64(30-targetLevel) / 29.0
							writeE = -Ec * (1.0 + ratio*1.0) // -Ec to -2*Ec
						}
					}

					// Ramp to write field
					diff := writeE - a.electricField
					step := rampRate * dt
					if math.Abs(diff) < step {
						a.electricField = writeE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}

					// Transition when field applied and held
					if a.manualPhaseTime > phaseDuration*0.4 && math.Abs(a.electricField-writeE) < 0.01*Emax {
						a.manualPhase = 3
						a.manualPhaseTime = 0
					}

				case 3: // HOLD_WRITE - return to zero, polarization persists
					step := rampRate * dt
					if math.Abs(a.electricField) < step {
						a.electricField = 0
					} else if a.electricField > 0 {
						a.electricField -= step
					} else {
						a.electricField += step
					}

					// Animation complete
					if a.manualPhaseTime > phaseDuration*0.3 && math.Abs(a.electricField) < 0.01*Emax {
						finalLevel := a.discreteLevel + 1
						levelError := finalLevel - targetLevel

						// Update calibration based on error (for next time)
						// This adaptive calibration improves accuracy over time
						if levelError != 0 && a.calibrated {
							if targetLevel > startLevel {
								// Adjust ascending calibration
								adjustment := float64(levelError) * Ec * 0.02
								a.calibrationUp[targetLevel-1] -= adjustment
							} else {
								// Adjust descending calibration
								adjustment := float64(levelError) * Ec * 0.02
								a.calibrationDown[targetLevel-1] += adjustment
							}
						}

						a.manualAnimating = false
						a.manualPhase = 0
						a.addLogEntry(fmt.Sprintf("→ Level %d (target %d)", finalLevel, targetLevel))
					}
				}
			}
			// If not animating, electric field is already set by slider in controls.go
		} else if a.autoMode {
			Emax := a.material.Ec * 2
			// Wrap phase to prevent floating-point precision loss over long times
			phase := math.Mod(2*math.Pi*a.frequency*a.simTime, 2*math.Pi)

			switch a.waveform {
			case WaveformSine:
				a.electricField = Emax * math.Sin(phase)
			case WaveformTriangle:
				p := phase / (2 * math.Pi)
				if p < 0.25 {
					a.electricField = Emax * (4 * p)
				} else if p < 0.75 {
					a.electricField = Emax * (2 - 4*p)
				} else {
					a.electricField = Emax * (4*p - 4)
				}
			case WaveformWriteReadDemo:
				// Correct ferroelectric write/read physics with RESET-AND-RETRY approach:
				//
				// PHYSICS: Hysteresis is PATH-DEPENDENT and NON-REVERSIBLE.
				// If you overshoot a target level, you CANNOT correct by applying less field
				// or opposite field (that's a different branch of the hysteresis loop).
				// You MUST reset to a known saturation state and apply precise programming pulse.
				//
				// Phase mapping:
				// 0 = RESET (saturate in opposite direction to target)
				// 1 = HOLD_RESET (return to zero - now at known remanent: level 1 or 30)
				// 2 = WRITE (apply calibrated field toward target)
				// 3 = HOLD_WRITE (return to zero, polarization persists)
				// 4 = READ (small sense pulse below Ec)
				// 5 = DISPLAY (show result, pick next target)

				a.wrdPhaseTimer += dt
				phaseDuration := 1.0 / a.frequency
				rampRate := 3.0 * Emax * a.frequency
				Ec := a.material.Ec

				targetLevel := a.wrdTargetLevel // 1-indexed
				startLevel := a.wrdStartLevel   // Captured at cycle start

				switch a.wrdPhase {
				case 0: // RESET - saturate in opposite direction to target
					var resetE float64
					if targetLevel > startLevel || targetLevel > 15 {
						// Going UP or target in upper half: first saturate negative (reach level 1)
						resetE = -2.0 * Ec
					} else {
						// Going DOWN or target in lower half: first saturate positive (reach level 30)
						resetE = 2.0 * Ec
					}

					// Ramp to reset field
					diff := resetE - a.electricField
					step := rampRate * dt
					if math.Abs(diff) < step {
						a.electricField = resetE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}

					// Transition when field reached and held briefly
					if a.wrdPhaseTimer > phaseDuration*0.25 && math.Abs(a.electricField-resetE) < 0.01*Emax {
						a.wrdPhase = 1
						a.wrdPhaseTimer = 0
					}

				case 1: // HOLD_RESET - return to zero (now at known remanent state)
					step := rampRate * dt
					if math.Abs(a.electricField) < step {
						a.electricField = 0
					} else if a.electricField > 0 {
						a.electricField -= step
					} else {
						a.electricField += step
					}

					// Now at known remanent state (level 1 or level 30)
					if a.wrdPhaseTimer > phaseDuration*0.15 && math.Abs(a.electricField) < 0.01*Emax {
						a.wrdPhase = 2
						a.wrdPhaseTimer = 0
					}

				case 2: // WRITE - apply calibrated field for target
					var writeE float64
					goingUp := targetLevel > startLevel || targetLevel > 15

					if goingUp {
						// Going UP from level 1: use ascending calibration
						writeE = a.calibrationUp[targetLevel-1] // 0-indexed array
						if writeE == 0 {
							// Fallback: interpolate based on target position
							ratio := float64(targetLevel-1) / 29.0
							writeE = Ec * (1.0 + ratio*1.0) // Ec to 2*Ec
						}
					} else {
						// Going DOWN from level 30: use descending calibration
						writeE = a.calibrationDown[targetLevel-1] // 0-indexed array
						if writeE == 0 {
							ratio := float64(30-targetLevel) / 29.0
							writeE = -Ec * (1.0 + ratio*1.0) // -Ec to -2*Ec
						}
					}
					a.wrdWriteE = writeE // Store for logging

					// Ramp to write field
					diff := writeE - a.electricField
					step := rampRate * dt
					if math.Abs(diff) < step {
						a.electricField = writeE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}

					// Transition when field applied and held
					if a.wrdPhaseTimer > phaseDuration*0.3 && math.Abs(a.electricField-writeE) < 0.01*Emax {
						a.wrdPhase = 3
						a.wrdPhaseTimer = 0
					}

				case 3: // HOLD_WRITE - return to zero, polarization persists
					step := rampRate * dt
					if math.Abs(a.electricField) < step {
						a.electricField = 0
					} else if a.electricField > 0 {
						a.electricField -= step
					} else {
						a.electricField += step
					}

					// Transition to READ phase
					if a.wrdPhaseTimer > phaseDuration*0.2 && math.Abs(a.electricField) < 0.01*Emax {
						a.wrdPhase = 4
						a.wrdPhaseTimer = 0
					}

				case 4: // READ phase - small sense pulse below Ec
					readE := Ec * 0.3 // Well below Ec - won't switch
					step := rampRate * 0.4 * dt
					diff := readE - a.electricField
					if math.Abs(diff) < step {
						a.electricField = readE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}
					// Capture read level and transition
					if a.wrdPhaseTimer > phaseDuration*0.3 {
						a.wrdReadLevel = a.discreteLevel + 1
						a.wrdPhase = 5
						a.wrdPhaseTimer = 0

						// Track Dr. Tour demo metrics
						a.wrdTotalWrites++
						// Success if within ±1 level (analog tolerance)
						levelError := a.wrdReadLevel - a.wrdTargetLevel
						if abs(levelError) <= 1 {
							a.wrdSuccessWrites++
						}

						// Update calibration based on error (adaptive learning)
						if levelError != 0 && a.calibrated {
							goingUp := a.wrdTargetLevel > a.wrdStartLevel || a.wrdTargetLevel > 15
							if goingUp {
								// Adjust ascending calibration
								adjustment := float64(levelError) * Ec * 0.02
								a.calibrationUp[a.wrdTargetLevel-1] -= adjustment
							} else {
								// Adjust descending calibration
								adjustment := float64(levelError) * Ec * 0.02
								a.calibrationDown[a.wrdTargetLevel-1] += adjustment
							}
						}

						// Add accumulated energy for this cycle (calculated from E·dP integration)
						a.wrdTotalEnergyfJ += a.wrdCycleEnergy
						a.wrdCycleEnergy = 0 // Reset for next cycle

						// Log this cycle for debugging
						if a.wrdDebugLog != nil {
							cycle := WriteReadCycle{
								CycleNum:    len(a.wrdDebugLog.Cycles) + 1,
								TargetLevel: a.wrdTargetLevel,
								StartLevel:  a.wrdStartLevel,
								ReadLevel:   a.wrdReadLevel,
								Success:     abs(a.wrdReadLevel-a.wrdTargetLevel) <= 1,
								Phases: []WriteReadPhase{
									{Phase: "RESET", EFieldPeak: a.wrdSaturateE / 1e8},
									{Phase: "WRITE", EFieldPeak: a.wrdWriteE / 1e8},
									{Phase: "READ", EFieldPeak: readE / 1e8, LevelEnd: a.wrdReadLevel},
								},
							}
							a.wrdDebugLog.Cycles = append(a.wrdDebugLog.Cycles, cycle)

							// Cap debug log to 100 cycles to prevent memory leak
							if len(a.wrdDebugLog.Cycles) > 100 {
								a.wrdDebugLog.Cycles = a.wrdDebugLog.Cycles[len(a.wrdDebugLog.Cycles)-100:]
							}

							// Save after every 5 cycles
							if len(a.wrdDebugLog.Cycles)%5 == 0 {
								go a.saveDebugLog()
							}
						}
					}

				case 5: // DISPLAY phase - return to zero, show result
					step := rampRate * 0.4 * dt
					if math.Abs(a.electricField) < step {
						a.electricField = 0
					} else if a.electricField > 0 {
						a.electricField -= step
					} else {
						a.electricField += step
					}
					// Transition to next cycle
					if a.wrdPhaseTimer > phaseDuration*0.6 {
						// Record start level for next cycle
						a.wrdStartLevel = a.discreteLevel + 1

						// Add comparison callout every 5 cycles
						if a.wrdTotalWrites > 0 && a.wrdTotalWrites%5 == 0 {
							fecimEnergy := a.wrdTotalEnergyfJ / 1000 // pJ
							// NOTE: 10M× is Dr. Tour's unverified claim. Peer-reviewed: 25-100× (Samsung Nature 2025)
							nandEquiv := fecimEnergy * 50            // 25-100× better (conservative: use 50)
							dramEquiv := fecimEnergy * 1000          // 1000× worse
							bitsStored := float64(a.wrdTotalWrites) * 4.91
							a.addLogEntry("━━ ENERGY COMPARISON ━━")
							a.addLogEntry(fmt.Sprintf("FeCIM: %.0f pJ total", fecimEnergy))
							a.addLogEntry(fmt.Sprintf("NAND:  %.0f pJ (50×!)", nandEquiv))
							a.addLogEntry(fmt.Sprintf("DRAM:  %.0f pJ (1000×)", dramEquiv))
							a.addLogEntry(fmt.Sprintf("Bits stored: %.0f (%.1f×binary)", bitsStored, 4.91))
							a.addLogEntry("━━━━━━━━━━━━━━━━━━━━━━")
						}

						// Milestone celebrations
						switch a.wrdTotalWrites {
						case 10:
							a.addLogEntry("★★ 10 ops! ~49 bits stored ★★")
						case 25:
							a.addLogEntry("★★★ 25 ops! ~123 bits stored ★★★")
						case 50:
							a.addLogEntry("★★★★ 50 ops! ~245 bits stored ★★★★")
							a.addLogEntry("Binary would need 245 cells!")
							a.addLogEntry("FeCIM: only 50 cells! (5× denser)")
						case 100:
							a.addLogEntry("★★★★★ 100 OPERATIONS! ★★★★★")
							a.addLogEntry("~491 bits in 100 FeCIM cells")
							a.addLogEntry("Binary: 491 cells needed!")
							successRate := float64(a.wrdSuccessWrites) / float64(a.wrdTotalWrites) * 100
							a.addLogEntry(fmt.Sprintf("Accuracy: %.0f%%", successRate))
						}

						// Pick new target - alternate between high and low
						if a.wrdTargetLevel > 15 {
							a.wrdTargetLevel = rand.Intn(12) + 2 // Low: 2-13 (avoid extremes)
						} else {
							a.wrdTargetLevel = rand.Intn(12) + 18 // High: 18-29 (avoid extremes)
						}
						a.wrdPhase = 0
						a.wrdPhaseTimer = 0
						a.wrdCycleEnergy = 0 // Reset energy accumulator for next cycle
					}
				}
			}
		}

		// Update physics
		prevP := a.polarization
		a.polarization = a.preisach.Update(a.electricField)
		a.normalizedP = a.preisach.NormalizedPolarization()
		a.discreteLevel = int(math.Round((a.normalizedP + 1) / 2 * 29))
		if a.discreteLevel < 0 {
			a.discreteLevel = 0
		}
		if a.discreteLevel > 29 {
			a.discreteLevel = 29
		}

		// Calculate energy: integral of E·dP ≈ |E| * |ΔP|
		// During write/read cycles, accumulate energy for the cycle (phases 0-4)
		if a.waveform == WaveformWriteReadDemo && a.wrdPhase >= 0 && a.wrdPhase <= 4 {
			deltaP := a.polarization - prevP
			// Energy per unit volume: E·dP in J/m³
			// Use actual cell dimensions from material
			cellVolume := a.material.Area * a.material.Thickness
			// Fallback if material doesn't have dimensions
			if cellVolume <= 0 {
				cellVolume = 2e-22 // Default: 100nm x 100nm x 20nm
			}
			energyJ := math.Abs(a.electricField * deltaP) * cellVolume
			energyfJ := energyJ * 1e15 // Convert J to fJ
			a.wrdCycleEnergy += energyfJ
		}

		// Record history
		a.eHistory = append(a.eHistory, a.electricField)
		a.pHistory = append(a.pHistory, a.polarization)
		if len(a.eHistory) > a.maxHistory {
			a.eHistory = a.eHistory[1:]
			a.pHistory = a.pHistory[1:]
		}

		// Copy data for UI update
		eField := a.electricField
		pol := a.polarization
		level := a.discreteLevel
		eHist := make([]float64, len(a.eHistory))
		pHist := make([]float64, len(a.pHistory))
		copy(eHist, a.eHistory)
		copy(pHist, a.pHistory)

		a.mu.Unlock()

		// Update UI (must be on main thread)
		a.updateUI(eField, pol, level, eHist, pHist)
	}
}

// updateUI updates all UI elements with the latest simulation data
func (a *App) updateUI(eField, pol float64, level int, eHist, pHist []float64) {
	fyne.Do(func() {
		// Update labels
		a.eFieldLabel.SetText(fmt.Sprintf("E-field: %.3f MV/cm", eField/1e8))
		a.pLabel.SetText(fmt.Sprintf("%.2f µC/cm²", pol*100))
		a.levelLabel.SetText(fmt.Sprintf("%d/30", level+1))

		// Update state descriptor
		var stateText string
		if level < 10 {
			stateText = "Negative P"
		} else if level > 19 {
			stateText = "Positive P"
		} else {
			stateText = "Intermediate"
		}
		if a.stateLabel != nil {
			a.stateLabel.SetText(stateText)
		}

		// Update wake-up/fatigue labels (Dr. Tour recommendation)
		cycles, degradation, wakeup := a.preisach.GetFatigueState()
		if a.cyclesLabel != nil {
			if cycles >= 1000000 {
				a.cyclesLabel.SetText(fmt.Sprintf("%.1fM", float64(cycles)/1e6))
			} else if cycles >= 1000 {
				a.cyclesLabel.SetText(fmt.Sprintf("%.1fK", float64(cycles)/1e3))
			} else {
				a.cyclesLabel.SetText(fmt.Sprintf("%d", cycles))
			}
		}
		if a.wakeupLabel != nil {
			a.wakeupLabel.SetText(fmt.Sprintf("%.1f%%", wakeup*100))
		}
		if a.fatigueLabel != nil {
			a.fatigueLabel.SetText(fmt.Sprintf("%.4f%%", degradation*100))
		}

		// Update WRITE/READ mode indicator based on E vs Ec
		isWrite := math.Abs(eField) > a.material.Ec
		a.modeIndicator.SetWrite(isWrite)
		a.modeIndicator.Refresh()

		// Update slider to match current E-field (only if not being manually controlled)
		// During Manual animation, the slider reflects the animated E-field
		// Normalize by Ec for display (-2 to +2 range)
		a.mu.RLock()
		shouldUpdateSlider := a.waveform != WaveformManual || a.manualAnimating
		a.mu.RUnlock()
		if shouldUpdateSlider {
			a.eFieldSlider.SetValue(eField / a.material.Ec)
		}

		// Update status and logging
		if a.paused {
			a.statusLabel.SetText("⏸ Paused")
		} else {
			a.mu.RLock()
			waveform := a.waveform
			wrdPhase := a.wrdPhase
			wrdTarget := a.wrdTargetLevel
			wrdRead := a.wrdReadLevel
			lastPhase := a.lastLogPhase
			wrdTotalWrites := a.wrdTotalWrites
			wrdSuccessWrites := a.wrdSuccessWrites
			wrdTotalEnergyfJ := a.wrdTotalEnergyfJ
			a.mu.RUnlock()

			switch waveform {
			case WaveformWriteReadDemo:
				var phaseStr string
				// Log phase transitions (6 phases: RESET, HOLD_RESET, WRITE, HOLD_WRITE, READ, DISPLAY)
				if wrdPhase != lastPhase {
					a.mu.Lock()
					a.lastLogPhase = wrdPhase
					switch wrdPhase {
					case 0:
						// RESET: Saturate in opposite direction
						direction := "-sat"
						if wrdTarget <= 15 {
							direction = "+sat"
						}
						a.addLogEntry(fmt.Sprintf("◆◆ RESET   | %s | prep", direction))
					case 1:
						// HOLD_RESET: Return to zero (known state)
						a.addLogEntry("░░ SETTLE  | E=0 | prep done")
					case 2:
						// WRITE: Apply calibrated field to reach target
						direction := "+"
						if wrdTarget <= 15 {
							direction = "-"
						}
						a.addLogEntry(fmt.Sprintf("▓▓ WRITE L%d | %sE>Ec | ~10fJ", wrdTarget, direction))
					case 3:
						// HOLD_WRITE: Return to zero, polarization persists
						a.addLogEntry(fmt.Sprintf("░░ HOLD L%d | E=0 | 0 fJ!", level+1))
					case 4:
						// READ: Non-destructive sense
						a.addLogEntry("▒▒ READ    | E<Ec | ~1fJ")
					case 5:
						// DISPLAY: Show result
						status := "✓ MATCH"
						if wrdRead != wrdTarget {
							diff := abs(wrdRead - wrdTarget)
							if diff == 1 {
								status = fmt.Sprintf("△ ±1 (got %d)", wrdRead)
							} else {
								status = fmt.Sprintf("✗ miss (got %d)", wrdRead)
							}
						}
						successRate := 0.0
						if wrdTotalWrites > 0 {
							successRate = float64(wrdSuccessWrites) / float64(wrdTotalWrites) * 100
						}
						a.addLogEntry(fmt.Sprintf("●● L%d %s [%.0f%% rate]", wrdTarget, status, successRate))
					}
					a.mu.Unlock()
				}

				// Enhanced status with energy metrics (using local copies from RLock above)
				energyTotal := wrdTotalEnergyfJ
				writeCount := wrdTotalWrites

				switch wrdPhase {
				case 0:
					direction := "-sat"
					if wrdTarget <= 15 {
						direction = "+sat"
					}
					phaseStr = fmt.Sprintf("◆ RESET | %s | preparing", direction)
				case 1:
					phaseStr = "░ SETTLE | E=0 | at known state"
				case 2:
					direction := "+"
					if wrdTarget <= 15 {
						direction = "-"
					}
					phaseStr = fmt.Sprintf("▓ WRITE L%d | %sE>Ec | ~10fJ", wrdTarget, direction)
				case 3:
					phaseStr = fmt.Sprintf("░ HOLD L%d | E=0 | ZERO POWER", level+1)
				case 4:
					phaseStr = fmt.Sprintf("▒ READ | Sense L%d | ~1fJ", level+1)
				case 5:
					successRate := 0.0
					if writeCount > 0 {
						successRate = float64(wrdSuccessWrites) / float64(writeCount) * 100
					}
					if wrdRead == wrdTarget {
						phaseStr = fmt.Sprintf("● L%d ✓ | Ops:%d | %.0f%% | %.0fpJ", wrdRead, writeCount, successRate, energyTotal/1000)
					} else {
						phaseStr = fmt.Sprintf("● L%d (want %d) | Ops:%d | %.0f%%", wrdRead, wrdTarget, writeCount, successRate)
					}
				}
				a.statusLabel.SetText(fmt.Sprintf("⚡ FeCIM Write/Read | %s", phaseStr))
			case WaveformManual:
				// Manual mode status with RESET-AND-RETRY physics
				a.mu.RLock()
				animating := a.manualAnimating
				manPhase := a.manualPhase
				manTarget := a.manualTargetLevel
				manStart := a.manualStartLevel
				a.mu.RUnlock()

				if animating {
					var phaseStr string
					switch manPhase {
					case 0:
						// RESET phase
						if manTarget > manStart {
							phaseStr = "RESET -sat..."
						} else {
							phaseStr = "RESET +sat..."
						}
					case 1:
						phaseStr = "SETTLE E=0..."
					case 2:
						phaseStr = fmt.Sprintf("WRITE → L%d...", manTarget)
					case 3:
						phaseStr = fmt.Sprintf("HOLD L%d...", level+1)
					default:
						phaseStr = fmt.Sprintf("Current: L%d", level+1)
					}
					a.statusLabel.SetText(fmt.Sprintf("TARGET L%d | %s", manTarget, phaseStr))
				} else {
					a.statusLabel.SetText(fmt.Sprintf("Manual L%d | Click level bar", level+1))
				}
			default:
				frac := a.preisach.GetSwitchedFraction() * 100
				a.statusLabel.SetText(fmt.Sprintf("● Running | t=%.2fs | Switched: %.1f%%", a.simTime, frac))
			}
		}

		// Update slide text based on current waveform
		a.slideText.SetText(a.getSlideText())

		// Update log text
		a.mu.RLock()
		logText := a.getLogText()
		a.mu.RUnlock()
		a.logText.SetText(logText)

		// Update plot
		a.plot.SetData(eHist, pHist, eField, pol)
		a.plot.Refresh()

		// Update level indicator
		a.levelIndicator.SetLevel(level)

		// Highlight target level during animations
		a.mu.RLock()
		currentWaveform := a.waveform
		currentWrdPhase := a.wrdPhase
		currentWrdTarget := a.wrdTargetLevel
		manualAnim := a.manualAnimating
		manualTarget := a.manualTargetLevel
		a.mu.RUnlock()

		if currentWaveform == WaveformWriteReadDemo {
			// Show target during phases 0-4 (RESET/SETTLE/WRITE/HOLD/READ)
			highlight := currentWrdPhase >= 0 && currentWrdPhase <= 4
			a.levelIndicator.SetTargetLevel(currentWrdTarget, highlight)
		} else if currentWaveform == WaveformManual && manualAnim {
			// Show target during Manual mode click animation
			a.levelIndicator.SetTargetLevel(manualTarget, true)
		} else {
			// Clear target highlight
			a.levelIndicator.SetTargetLevel(0, false)
		}

		a.levelIndicator.Refresh()

		// Update cell visualizer
		a.cellViz.SetLevel(level)
		a.cellViz.Refresh()
	})
}

// calibrateLevels performs a calibration sweep to map field→level relationship.
// This mimics how real ferroelectric memory controllers characterize each device
// and build lookup tables for programming. Called at startup and when material changes.
// MUST be called with a.mu held.
func (a *App) calibrateLevels() {
	if a.preisach == nil || a.material == nil {
		return
	}

	Ec := a.material.Ec
	Emax := 2.5 * Ec // Go slightly beyond saturation

	// Save current Preisach state (we'll restore after calibration)
	// Note: We can't fully save/restore Preisach state, so we'll reset after

	// Calibrate ASCENDING (from negative saturation to positive)
	// First, saturate negative
	for i := 0; i < 100; i++ {
		a.preisach.Update(-Emax)
	}
	a.preisach.Update(0) // Return to zero (remanent state)

	// Now sweep up and record field needed for each level
	lastLevel := 0
	for e := 0.0; e <= Emax; e += Ec * 0.02 { // Fine steps (2% of Ec)
		a.preisach.Update(e)
		p := a.preisach.Update(0) // Check remanent after removing field
		normalizedP := p / a.material.Ps
		level := int(math.Round((normalizedP + 1) / 2 * 29))
		if level < 0 {
			level = 0
		}
		if level > 29 {
			level = 29
		}

		// Record the field that first achieved this level
		if level > lastLevel && level < 30 {
			a.calibrationUp[level] = e
			lastLevel = level
		}

		// Re-apply field for next iteration (continue sweep)
		a.preisach.Update(e)
	}

	// Fill any gaps (use interpolation)
	for i := 1; i < 30; i++ {
		if a.calibrationUp[i] == 0 && i > 0 {
			a.calibrationUp[i] = a.calibrationUp[i-1] + Ec*0.05
		}
	}

	// Calibrate DESCENDING (from positive saturation to negative)
	// First, saturate positive
	for i := 0; i < 100; i++ {
		a.preisach.Update(Emax)
	}
	a.preisach.Update(0) // Return to zero (remanent state)

	lastLevel = 29
	for e := 0.0; e >= -Emax; e -= Ec * 0.02 { // Fine steps (negative direction)
		a.preisach.Update(e)
		p := a.preisach.Update(0)
		normalizedP := p / a.material.Ps
		level := int(math.Round((normalizedP + 1) / 2 * 29))
		if level < 0 {
			level = 0
		}
		if level > 29 {
			level = 29
		}

		// Record the field that first achieved this level (going down)
		if level < lastLevel && level >= 0 {
			a.calibrationDown[level] = e // Negative field
			lastLevel = level
		}

		// Re-apply field for next iteration
		a.preisach.Update(e)
	}

	// Fill gaps (descending)
	for i := 28; i >= 0; i-- {
		if a.calibrationDown[i] == 0 && i < 29 {
			a.calibrationDown[i] = a.calibrationDown[i+1] - Ec*0.05
		}
	}

	// Reset Preisach to neutral state after calibration
	a.preisach.Reset()
	a.electricField = 0
	a.polarization = 0
	a.calibrated = true

	log.Info("Level calibration complete for material: %s", a.material.Name)
}

// getCalibratedWriteField returns the calibrated E-field for a target level.
// DEPRECATED: The RESET-AND-RETRY approach doesn't need continuous feedback.
// This function is kept for backward compatibility but the new physics implementation
// directly accesses calibrationUp/calibrationDown arrays.
//
// The correct approach for ferroelectric programming:
// 1. RESET to known saturation (opposite direction to target)
// 2. Return to E=0 (now at known remanent: level 1 or 30)
// 3. Apply single calibrated pulse from calibrationUp or calibrationDown
// 4. Return to E=0 (polarization persists at target level)
//
// If target missed, record error and adjust calibration for next time.
// Do NOT try to "correct" by applying opposite field - that's physically wrong.
func (a *App) getCalibratedWriteField(currentLevel, targetLevel, startLevel int) float64 {
	Ec := a.material.Ec

	// If already at target, no field needed
	if currentLevel == targetLevel {
		return 0
	}

	// Determine direction based on start level
	goingUp := targetLevel > startLevel

	if goingUp {
		// Use ascending calibration (from level 1)
		field := a.calibrationUp[targetLevel-1] // 0-indexed array
		if field == 0 {
			// Fallback: interpolate
			ratio := float64(targetLevel-1) / 29.0
			field = Ec * (1.0 + ratio*1.0)
		}
		return field
	} else {
		// Use descending calibration (from level 30)
		field := a.calibrationDown[targetLevel-1] // 0-indexed array
		if field == 0 {
			// Fallback: interpolate
			ratio := float64(30-targetLevel) / 29.0
			field = -Ec * (1.0 + ratio*1.0)
		}
		return field
	}
}
