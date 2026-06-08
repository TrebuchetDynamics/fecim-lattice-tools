package progress

import "fyne.io/fyne/v2/data/binding"

// ProgressBindings exposes a Progress tracker as Fyne data bindings.
//
// These bindings let long-running FeCIM workflows update labels and progress
// bars through binding-aware widgets instead of scattering direct widget
// mutation across goroutine callbacks.
type ProgressBindings struct {
	Operation     binding.String
	Phase         binding.String
	Detail        binding.String
	State         binding.String
	StatusLine    binding.String
	Fraction      binding.Float
	Percent       binding.Float
	Indeterminate binding.Bool
}

// NewProgressBindings creates bindings for p and keeps them synchronized with
// subsequent progress updates, completion, and failure events.
func NewProgressBindings(p *Progress) *ProgressBindings {
	b := &ProgressBindings{
		Operation:     binding.NewString(),
		Phase:         binding.NewString(),
		Detail:        binding.NewString(),
		State:         binding.NewString(),
		StatusLine:    binding.NewString(),
		Fraction:      binding.NewFloat(),
		Percent:       binding.NewFloat(),
		Indeterminate: binding.NewBool(),
	}

	b.Sync(p)
	p.OnProgress(func(progress *Progress) {
		b.Sync(progress)
	})
	p.OnComplete(func(progress *Progress) {
		b.Sync(progress)
	})
	p.OnError(func(progress *Progress, _ error) {
		b.Sync(progress)
	})

	return b
}

// Sync copies the current Progress snapshot into the bindings.
func (b *ProgressBindings) Sync(p *Progress) {
	_ = b.Operation.Set(p.Operation())
	_ = b.Phase.Set(p.Phase())
	_ = b.Detail.Set(p.Detail())
	_ = b.State.Set(p.State().String())
	_ = b.StatusLine.Set(p.StatusLine())
	_ = b.Fraction.Set(p.Fraction())
	_ = b.Percent.Set(p.Percent())
	_ = b.Indeterminate.Set(p.IsIndeterminate())
}
