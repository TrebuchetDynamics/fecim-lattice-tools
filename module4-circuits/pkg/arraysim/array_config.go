package arraysim

import sharedphysics "fecim-lattice-tools/shared/physics"

// ArrayConfig captures common knobs for array analyses/simulations.
type ArrayConfig struct {
	Rows         int
	Cols         int
	ReadVoltageV float64
	CouplingMode CouplingMode
	Geometry     sharedphysics.CellGeometry
	Wire         WireParams
	Boundary     BoundaryParams
	Sense        SenseChain
	Material     *sharedphysics.HZOMaterial
}
