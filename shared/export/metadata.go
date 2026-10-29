package export

import "fecim-lattice-tools/shared/export/metadata"

// Compatibility aliases retain existing native export bytes and public names.
type SimulationMetadata = metadata.SimulationMetadata

func DefaultSimulationMetadata() SimulationMetadata { return metadata.DefaultSimulationMetadata() }
func SimulationDisclaimer() string                  { return metadata.SimulationDisclaimer() }
func SPICEDisclaimer() string                       { return metadata.SPICEDisclaimer() }
