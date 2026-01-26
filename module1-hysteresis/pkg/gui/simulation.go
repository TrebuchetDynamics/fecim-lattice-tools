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
			if a.manualAnimating {
				// Animation in progress - run saturate→settle→hold sequence
				Emax := a.material.Ec * 2
				Ec := a.material.Ec
				phaseDuration := 0.8 / a.frequency
				rampRate := 3.0 * Emax * a.frequency

				a.manualPhaseTime += dt

				// Calculate target fields same as Write/Read demo
				targetNormP := (float64(a.manualTargetLevel) - 15.5) / 14.5
				var saturateE, settleE float64
				if targetNormP >= 0 {
					saturateE = Emax
					settleRatio := 1.0 - (float64(a.manualTargetLevel-16) / 14.0)
					settleE = -Ec * (0.5 + settleRatio*1.0)
				} else {
					saturateE = -Emax
					settleRatio := float64(a.manualTargetLevel-1) / 14.0
					settleE = Ec * (0.5 + settleRatio*1.0)
				}

				switch a.manualPhase {
				case 1: // SATURATE - ramp to saturation
					diff := saturateE - a.electricField
					step := rampRate * dt
					if math.Abs(diff) < step {
						a.electricField = saturateE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}
					if a.manualPhaseTime > phaseDuration*0.5 && math.Abs(a.electricField-saturateE) < 0.01*Emax {
						a.manualPhase = 2
						a.manualPhaseTime = 0
					}

				case 2: // SETTLE - create minor loop
					diff := settleE - a.electricField
					step := rampRate * 0.7 * dt
					if math.Abs(diff) < step {
						a.electricField = settleE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}
					if a.manualPhaseTime > phaseDuration*0.4 && math.Abs(a.electricField-settleE) < 0.01*Emax {
						a.manualPhase = 3
						a.manualPhaseTime = 0
					}

				case 3: // HOLD - return to zero
					step := rampRate * dt
					if math.Abs(a.electricField) < step {
						a.electricField = 0
					} else if a.electricField > 0 {
						a.electricField -= step
					} else {
						a.electricField += step
					}
					if a.manualPhaseTime > phaseDuration*0.3 && math.Abs(a.electricField) < 0.01*Emax {
						a.manualAnimating = false
						a.manualPhase = 0
						a.addLogEntry(fmt.Sprintf("HOLD Level %d ✓", a.discreteLevel+1))
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
				// IMPROVED Write/Read Demo with correct ferroelectric physics
				//
				// Key insight: To write ANY level, you must first SATURATE then SETTLE
				// 1. SATURATE: Apply large |E| >> Ec to fully switch all hysterons
				// 2. SETTLE: Apply opposite field to create minor loop and set level
				// 3. HOLD: E = 0, polarization persists (non-volatile!)
				// 4. READ: Small sense pulse |E| < Ec, doesn't disturb state
				//
				// Phase mapping:
				// 0 = SATURATE (ramp to ±Emax)
				// 1 = SETTLE (ramp to settling field to set target level)
				// 2 = HOLD (return to zero)
				// 3 = READ (small sense pulse)
				// 4 = DISPLAY (show result, pick next target)

				a.wrdPhaseTimer += dt
				phaseDuration := 1.0 / a.frequency
				rampRate := 2.5 * Emax * a.frequency
				Ec := a.material.Ec

				// Calculate target normalized polarization (-1 to +1)
				// Level 1 = -1 (negative saturation)
				// Level 15-16 = ~0 (intermediate)
				// Level 30 = +1 (positive saturation)
				targetNormP := (float64(a.wrdTargetLevel) - 15.5) / 14.5

				// Determine saturation direction and settle field
				// For positive targets: saturate positive, settle negative
				// For negative targets: saturate negative, settle positive
				if targetNormP >= 0 {
					a.wrdSaturateE = Emax // Positive saturation
					// Settle field: partial negative field to create minor loop
					// More negative field → lower final level
					// The relationship is approximately: final P ≈ Ps * tanh((E_settle + Ec) / delta)
					// For level 30: settle at ~0 (stay at positive saturation)
					// For level 16: settle at ~ -Ec (intermediate)
					settleRatio := 1.0 - (float64(a.wrdTargetLevel-16) / 14.0) // 0 to 1 for levels 16-30
					a.wrdSettleE = -Ec * (0.5 + settleRatio*1.0)              // Range: -0.5*Ec to -1.5*Ec
				} else {
					a.wrdSaturateE = -Emax // Negative saturation
					// For negative targets, settle with positive field
					// For level 1: settle at ~0 (stay at negative saturation)
					// For level 15: settle at ~ +Ec (intermediate)
					settleRatio := float64(a.wrdTargetLevel-1) / 14.0 // 0 to 1 for levels 1-15
					a.wrdSettleE = Ec * (0.5 + settleRatio*1.0)       // Range: 0.5*Ec to 1.5*Ec
				}

				switch a.wrdPhase {
				case 0: // SATURATE phase - ramp to saturation voltage (±Emax)
					diff := a.wrdSaturateE - a.electricField
					step := rampRate * dt
					if math.Abs(diff) < step {
						a.electricField = a.wrdSaturateE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}
					// Transition when we've reached saturation and held briefly
					if a.wrdPhaseTimer > phaseDuration*0.6 && math.Abs(a.electricField-a.wrdSaturateE) < 0.01*Emax {
						a.wrdPhase = 1
						a.wrdPhaseTimer = 0
					}

				case 1: // SETTLE phase - ramp to settling field (creates minor loop)
					diff := a.wrdSettleE - a.electricField
					step := rampRate * 0.7 * dt // Slower for precision
					if math.Abs(diff) < step {
						a.electricField = a.wrdSettleE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}
					// Transition when we've reached settle field
					if a.wrdPhaseTimer > phaseDuration*0.5 && math.Abs(a.electricField-a.wrdSettleE) < 0.01*Emax {
						a.wrdPhase = 2
						a.wrdPhaseTimer = 0
					}

				case 2: // HOLD phase - return to zero (polarization persists!)
					step := rampRate * dt
					if math.Abs(a.electricField) < step {
						a.electricField = 0
					} else if a.electricField > 0 {
						a.electricField -= step
					} else {
						a.electricField += step
					}
					// Transition when at zero
					if a.wrdPhaseTimer > phaseDuration*0.5 && math.Abs(a.electricField) < 0.01*Emax {
						a.wrdPhase = 3
						a.wrdPhaseTimer = 0
					}

				case 3: // READ phase - small sense pulse below Ec
					readE := Ec * 0.3 // Well below Ec - won't switch
					if a.wrdSaturateE < 0 {
						readE = -readE // Match polarity of written state
					}
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
					if a.wrdPhaseTimer > phaseDuration*0.4 {
						a.wrdReadLevel = a.discreteLevel + 1
						a.wrdPhase = 4
						a.wrdPhaseTimer = 0

						// Track Dr. Tour demo metrics
						a.wrdTotalWrites++
						// Success if within ±1 level (analog tolerance)
						if abs(a.wrdReadLevel-a.wrdTargetLevel) <= 1 {
							a.wrdSuccessWrites++
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
									{Phase: "SATURATE", EFieldPeak: a.wrdSaturateE / 1e8},
									{Phase: "SETTLE", EFieldEnd: a.wrdSettleE / 1e8},
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

				case 4: // DISPLAY phase - return to zero, show result
					step := rampRate * 0.4 * dt
					if math.Abs(a.electricField) < step {
						a.electricField = 0
					} else if a.electricField > 0 {
						a.electricField -= step
					} else {
						a.electricField += step
					}
					// Transition to next cycle
					if a.wrdPhaseTimer > phaseDuration*0.8 {
						// Record start level for next cycle
						a.wrdStartLevel = a.discreteLevel + 1

						// Add comparison callout every 5 cycles
						if a.wrdTotalWrites > 0 && a.wrdTotalWrites%5 == 0 {
							fecimEnergy := a.wrdTotalEnergyfJ / 1000 // pJ
							nandEquiv := fecimEnergy * 10000000      // 10M× worse
							dramEquiv := fecimEnergy * 1000          // 1000× worse
							bitsStored := float64(a.wrdTotalWrites) * 4.91
							a.addLogEntry("━━ ENERGY COMPARISON ━━")
							a.addLogEntry(fmt.Sprintf("FeCIM: %.0f pJ total", fecimEnergy))
							a.addLogEntry(fmt.Sprintf("NAND:  %.0f pJ (10M×!)", nandEquiv))
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
		// During write/read cycles, accumulate energy for the cycle
		if a.waveform == WaveformWriteReadDemo && a.wrdPhase >= 0 && a.wrdPhase <= 3 {
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
		a.pLabel.SetText(fmt.Sprintf("P: %.2f µC/cm²", pol*100))
		a.levelLabel.SetText(fmt.Sprintf("Level: %d/30", level+1))

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

		// Update slider position for auto modes
		if a.autoMode {
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
				// Log phase transitions (now 5 phases: SATURATE, SETTLE, HOLD, READ, DISPLAY)
				if wrdPhase != lastPhase {
					a.mu.Lock()
					a.lastLogPhase = wrdPhase
					switch wrdPhase {
					case 0:
						// SATURATE: Show domain saturation with energy
						a.addLogEntry(fmt.Sprintf("▓▓ WRITE L%d | E=2Ec | ~10fJ", wrdTarget))
					case 1:
						// SETTLE: Show minor loop formation
						polarization := float64(wrdTarget-15) / 15.0 * 100 // % of Pr
						a.addLogEntry(fmt.Sprintf("██ SETTLE  | P→%.0f%% Pr", polarization))
					case 2:
						// HOLD: Emphasize zero-power retention
						a.addLogEntry(fmt.Sprintf("░░ HOLD L%d | E=0 | 0 fJ!", level+1))
					case 3:
						// READ: Show non-destructive sense
						a.addLogEntry("▒▒ READ    | E<Ec | ~1fJ")
					case 4:
						// RESULT: Show success/tolerance and running stats
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
					phaseStr = fmt.Sprintf("▓ WRITE L%d | E=2×Ec | ~10fJ switching", wrdTarget)
				case 1:
					polarization := float64(wrdTarget-15) / 15.0 * 100
					phaseStr = fmt.Sprintf("█ SETTLE L%d | P→%.0f%% of Pr", wrdTarget, polarization)
				case 2:
					phaseStr = fmt.Sprintf("░ HOLD L%d | E=0 | ZERO POWER", level+1)
				case 3:
					phaseStr = fmt.Sprintf("▒ READ | Sense L%d | ~1fJ", level+1)
				case 4:
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
				// Manual mode status
				a.mu.RLock()
				animating := a.manualAnimating
				manPhase := a.manualPhase
				manTarget := a.manualTargetLevel
				a.mu.RUnlock()

				if animating {
					var phaseStr string
					switch manPhase {
					case 1:
						phaseStr = fmt.Sprintf("SATURATING → L%d...", manTarget)
					case 2:
						phaseStr = fmt.Sprintf("SETTLING → L%d...", manTarget)
					case 3:
						phaseStr = fmt.Sprintf("HOLDING L%d...", level+1)
					default:
						phaseStr = fmt.Sprintf("Current: L%d", level+1)
					}
					a.statusLabel.SetText(fmt.Sprintf("🎯 Manual (animating) | %s", phaseStr))
				} else {
					a.statusLabel.SetText(fmt.Sprintf("✋ Manual | L%d | Click level bar or use slider", level+1))
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
			// Show target during phases 0-3 (SATURATE through READ)
			highlight := currentWrdPhase >= 0 && currentWrdPhase <= 3
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
