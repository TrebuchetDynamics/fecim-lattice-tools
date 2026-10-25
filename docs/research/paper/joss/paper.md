---
title: 'FeCIM Lattice Tools: A reproducible simulation-only workbench for ferroelectric compute-in-memory'
tags:
  - Go
  - ferroelectricity
  - compute-in-memory
  - reproducible research
authors:
  - name: Juan Tamez
    affiliation: 1
affiliations:
  - name: Trebuchet Dynamics, Monterrey, Mexico
    index: 1
date: 25 October 2026
bibliography: paper.bib
---

> Working draft — not submitted or accepted. Research-use evidence, complete AI
> disclosure, author approval, and publication metadata remain release gates.

# Summary

FeCIM Lattice Tools is an open-source, simulation-only workbench for exploring
ferroelectric compute-in-memory systems. These systems seek to perform
calculations close to stored information rather than repeatedly moving it
between memory and a processor. Their behavior depends on assumptions about
memory devices, arrays, and supporting circuits. The workbench makes those
assumptions inspectable through local project files, parameter sweeps, and
reports. A Go command-line interface supports repeatable studies, while a Fyne
desktop interface provides interactive workspaces for model exploration.

The software produces simulation estimates, not measurements of fabricated
hardware. Its contribution is a reviewable computational workflow connecting
explicit assumptions to stored results. It is not a foundry process design kit
or a predictor certified against a particular commercial device.

# Statement of need

Researchers exploring compute-in-memory architectures must distinguish device
properties established by measurements from convenient simulation choices.
Changes in array size, conductance discretization, or converter resolution can
alter several estimated system metrics at once. Without a recorded project
definition and model identity, a comparison is difficult to reproduce or audit.

The intended users are device and circuit researchers, graduate students, and
engineers performing exploratory pre-silicon studies. FeCIM exposes a finite
design space with explicit objectives and feasibility constraints, retaining
the evaluated design and the assumptions behind its metrics. Evidence labels
distinguish literature-backed, experiment-calibrated, simulation-default, and
derived values. The default of 30 conductance levels is configurable educational
discretization; it is not a claim that a measured device supports 30 reliable
states. Evidence labels describe recorded provenance, not automatic certification
that a source or calibration is scientifically sufficient.

# State of the field

CrossSim is a Python-based analog crossbar simulator for neural networks and
linear algebra. It models non-idealities including programming errors, drift,
read noise, converter precision, and parasitic resistance [@crosssim].
IBM's Analog Hardware Acceleration Kit provides PyTorch integration and an
analog-device simulator for training and inference studies [@aihwkit].
These projects offer substantial capabilities that a small workbench should
not claim to replace.

FeCIM's rationale for a separate project is a local Go workflow combining
ferroelectric educational workspaces with explicit study bundles, content-based
run identity, feasibility analysis, and plain-file reports. This is a workflow
and integration choice, not evidence of superior simulation accuracy or speed.
Where the research question is primarily analog neural-network training or
detailed crossbar accuracy, established specialist tools may be the more
appropriate starting point. Cross-tool comparisons require aligned models,
units, input conditions, and documented tolerances; this manuscript does not
claim general numerical equivalence with either package.

# Software design

A study is a directory containing project, design, and sweep YAML files.
Project loading rejects unknown fields and validates declared local inputs,
including their containment and optional hashes. A finite Cartesian expansion
makes the set of evaluated points explicit and bounded. This favors inspectable
small studies over an opaque search service or an unrestricted optimizer.

Each run identity hashes the resolved design, seed, model and evaluator
versions, schema identity, and resolved input references. Bounded workers
evaluate points independently, and the runner records results in deterministic
order. Existing completed runs are verified before reuse. Runtime information
belongs to provenance rather than being allowed to perturb deterministic
report content. Model and evaluator version discipline remains necessary:
content addressing alone cannot detect an unrecorded change in scientific
assumptions.

The current HZO system evaluator composes simplified area, energy, latency,
and peripheral signal estimates. It does not run a measured-device
characterization or dynamically calibrate the system evaluator from a P–E
loop. A seed is recorded in run identity, although this evaluator currently
does not use stochastic sampling. Separating the evaluator from project
loading and storage permits these limitations to remain visible without
changing the experiment contract.

Model-domain failures remain stored outcomes rather than disappearing from
the design space. Analysis distinguishes successful evaluation from satisfying
constraints and computes Pareto status only relative to the configured
objective directions. JSON, CSV, and Markdown reports can be regenerated from
stored runs without reevaluating the models. This file-based design avoids
requiring a database or remote service for ordinary studies.

# Research impact statement

The repository provides a runnable HZO starter study and automated checks for
worker-count determinism, immutable-run reuse, input validation, failed-point
preservation, and deterministic reporting. Conservation-law checks include
Kirchhoff-law checks for crossbar computations, alongside numerical regressions.
These are computational correctness checks, not validation against measurements.
The corrected reference runner currently reports **0 experimental validations**.
These materials support inspection of the computational workflow. The eight-run example's feasible and Pareto
counts are a reproducibility demonstration, not a hardware performance result
or a scientific optimum.

This draft does not establish adoption by independent research groups,
publications enabled by the workbench, or a documented developer research
study beyond the starter example. Actual research-use evidence must be supplied
and assessed before submission. A passing regression suite is not a substitute
for that evidence or for independent experimental validation. In particular,
the explicitly synthetic Park-context shape fixture cannot demonstrate agreement
with measured HZO behavior. Its derived presets and reports are regression-only;
the separate unverified Park scalar JSON and disputed BTO extraction are excluded,
with original samples retained rather than smoothed. Experimental calibration, held-out validation, and
traceable extraction remain necessary before quantitative device claims.

# AI usage disclosure

This manuscript and its accompanying publication-preparation documents were
drafted with generative-AI assistance through Pi. Assistance included repository
inspection, documentation drafting, test scaffolding, and identification of
evidence gaps. This is not yet a complete disclosure for the software's
development history. The author must reconcile tools, model names and versions,
and assistance scope across code, tests, documentation, and manuscript before
submission. No human-review or design-decision attestation is asserted by this
AI-assisted draft; the human author must personally review, edit, validate,
and approve the submitted materials.

# Acknowledgements

Funding, relevant financial interests, and other acknowledgements require author
confirmation before submission. No external sponsorship, institutional
endorsement, or collaboration with a device company is asserted here.

# References
