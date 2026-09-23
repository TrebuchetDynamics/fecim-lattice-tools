package experiment

import "fecim-lattice-tools/workbench/project"

func Expand(bundle project.Bundle) ([]DesignPoint, error) {
	points, err := project.Expand(bundle)
	if err != nil {
		return nil, err
	}
	out := make([]DesignPoint, len(points))
	for i, point := range points {
		out[i] = DesignPoint{Index: point.Index, Design: point.Design, Seed: point.Seed, ModelVersion: bundle.Project.ModelVersion}
	}
	return out, nil
}
