package ferroelectric

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

func TestConcurrentMayergoyzUpdate(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)

	var wg sync.WaitGroup
	var panicCount atomic.Int32
	var nanCount atomic.Int32
	var infCount atomic.Int32

	goroutines := 10
	iterationsPerGoroutine := 100

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(seed int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicCount.Add(1)
				}
			}()

			rng := rand.New(rand.NewSource(int64(seed)))
			Emax := material.Ec * 2.0

			for i := 0; i < iterationsPerGoroutine; i++ {
				E := (rng.Float64()*2 - 1) * Emax
				P := model.Update(E)

				if math.IsNaN(P) {
					nanCount.Add(1)
				}
				if math.IsInf(P, 0) {
					infCount.Add(1)
				}
			}
		}(g)
	}

	wg.Wait()

	if panicCount.Load() > 0 {
		t.Errorf("Model panicked %d times under concurrent access", panicCount.Load())
	}

	if nanCount.Load() > 0 {
		t.Errorf("Model produced %d NaN values under concurrent access", nanCount.Load())
	}

	if infCount.Load() > 0 {
		t.Errorf("Model produced %d Inf values under concurrent access", infCount.Load())
	}
}

func TestConcurrentReadWhileUpdating(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)

	var wg sync.WaitGroup
	done := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		Emax := material.Ec * 2.0
		for i := 0; i < 1000; i++ {
			select {
			case <-done:
				return
			default:
				E := Emax * math.Sin(float64(i)*0.1)
				model.Update(E)
			}
		}
		close(done)
	}()

	var nanCount atomic.Int32
	for r := 0; r < 5; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					P := model.Polarization()
					if math.IsNaN(P) {
						nanCount.Add(1)
					}
				}
			}
		}()
	}

	wg.Wait()

	if nanCount.Load() > 0 {
		t.Errorf("Got %d NaN values while reading during updates", nanCount.Load())
	}
}

func TestConcurrentLoopGeneration(t *testing.T) {
	material := DefaultHZO()

	var wg sync.WaitGroup
	results := make(chan bool, 10)

	for g := 0; g < 5; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					results <- false
				}
			}()

			model := NewMayergoyzPreisach(material, 30+id*5)
			Emax := material.Ec * 2.0

			E, P := model.GetHysteresisLoop(Emax, 100+id*50)

			valid := true
			for i := range E {
				if math.IsNaN(E[i]) || math.IsNaN(P[i]) {
					valid = false
					break
				}
				if math.IsInf(E[i], 0) || math.IsInf(P[i], 0) {
					valid = false
					break
				}
			}
			results <- valid
		}(g)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	failures := 0
	for result := range results {
		if !result {
			failures++
		}
	}

	if failures > 0 {
		t.Errorf("%d concurrent loop generations failed", failures)
	}
}

func TestRaceConditionFieldHistory(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)

	var wg sync.WaitGroup

	Emax := material.Ec * 2.0

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			model.Update(Emax * float64(i%10) / 10)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			model.Update(-Emax * float64(i%10) / 10)
		}
	}()

	wg.Wait()

	P := model.Polarization()
	if math.IsNaN(P) {
		t.Error("Final polarization is NaN after concurrent updates")
	}
	if math.IsInf(P, 0) {
		t.Error("Final polarization is Inf after concurrent updates")
	}
}

func TestSimplePreisachConcurrency(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	var wg sync.WaitGroup
	var nanCount atomic.Int32

	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func(seed int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(int64(seed)))
			Emax := material.Ec * 2.0

			for i := 0; i < 100; i++ {
				E := (rng.Float64()*2 - 1) * Emax
				P := model.Update(E)
				if math.IsNaN(P) {
					nanCount.Add(1)
				}
			}
		}(g)
	}

	wg.Wait()

	if nanCount.Load() > 0 {
		t.Logf("SimplePreisach produced %d NaN values under concurrent access", nanCount.Load())
	}
}

func TestResetDuringUpdate(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)

	var wg sync.WaitGroup
	done := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		Emax := material.Ec * 2.0
		for i := 0; i < 500; i++ {
			select {
			case <-done:
				return
			default:
				model.Update(Emax * math.Sin(float64(i)*0.2))
			}
		}
		close(done)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				model.Reset()
			}
		}
	}()

	wg.Wait()

	P := model.Polarization()
	if math.IsNaN(P) {
		t.Error("Polarization is NaN after reset during update")
	}
}

func BenchmarkConcurrentUpdates(b *testing.B) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)
	Emax := material.Ec * 2.0

	b.RunParallel(func(pb *testing.PB) {
		rng := rand.New(rand.NewSource(42))
		for pb.Next() {
			E := (rng.Float64()*2 - 1) * Emax
			model.Update(E)
		}
	})
}

func TestParallelMaterialCreation(t *testing.T) {
	var wg sync.WaitGroup
	materialFuncs := []func() *HZOMaterial{
		DefaultHZO,
		FeCIMMaterial,
		FeCIMMaterialTarget,
		LiteratureSuperlattice,
		CryogenicHZO,
	}

	results := make(chan error, len(materialFuncs)*5)

	for _, getMaterial := range materialFuncs {
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(fn func() *HZOMaterial) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						results <- fmt.Errorf("panic: %v", r)
					}
				}()

				material := fn()
				model := NewMayergoyzPreisach(material, 30)
				E, P := model.GetHysteresisLoop(material.Ec*2, 50)

				if len(E) == 0 || len(P) == 0 {
					results <- fmt.Errorf("empty loop for %s", material.Name)
				}
			}(getMaterial)
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for err := range results {
		t.Error(err)
	}
}
