const modules = [
  [
    "Hysteresis Lab",
    "Explore Preisach and Landau-Khalatnikov ferroelectric switching with material presets for HZO, BTO, and PZT.",
  ],
  [
    "Crossbar Array",
    "Simulate matrix-vector multiply with quantized conductance levels, IR drop, sneak paths, and drift.",
  ],
  [
    "MNIST Pipeline",
    "Follow the full CIM inference path from input quantization through array computation and accuracy reporting.",
  ],
  [
    "Peripheral Circuits",
    "Inspect DAC, ADC, TIA, read-path, and ISPP write-control behavior around ferroelectric memory cells.",
  ],
  [
    "Technology Comparison",
    "Compare assumptions, operating points, and metrics across memory and compute-in-memory technologies.",
  ],
  [
    "EDA Export",
    "Generate SPICE, Verilog, Liberty, DEF, and LEF artifacts for education-stage OpenLane experiments.",
  ],
  [
    "Integrated Docs",
    "Keep physics explanations, references, module guides, and honesty notes connected to the simulator.",
  ],
];
export function Home() {
  return (
    <>
      <section className="hero">
        <div className="hero__content">
          <p className="eyebrow">Ferroelectric compute-in-memory</p>
          <h1>A lattice-scale playground for device-to-system ideas.</h1>
          <p>
            FeCIM Lattice Tools connects the physics of switching, the behavior
            of arrays, and the path from a material model to a computing
            experiment.
          </p>
          <p className="honesty">
            Education-phase simulation — material presets and 30 default
            conductance levels are educational assumptions, not validated device
            measurements.
          </p>
          <div className="actions">
            <a className="button primary" href="/hysteresis">
              Open Hysteresis lab
            </a>
            <a className="button" href="#modules">
              Explore the modules
            </a>
            <a className="button" href="#start">
              Run locally
            </a>
          </div>
        </div>
      </section>
      <section className="stats" aria-label="Project highlights">
        {[
          ["7", "integrated modules"],
          ["30", "default conductance levels"],
          ["MIT", "open-source license"],
        ].map(([value, label]) => (
          <article key={label}>
            <strong>{value}</strong>
            <span>{label}</span>
          </article>
        ))}
      </section>
      <section className="section">
        {/* Focus the introduction, not the viewport-tall grid, on fragment navigation. */}
        <header id="modules">
          <p className="eyebrow">Integrated workflow</p>
          <h2>Seven modules, one simulation story.</h2>
          <p>
            Browser candidates provide Hysteresis, bounded research studies,
            Crossbar, Circuits, Comparison and Documentation. MNIST inference and EDA generation require the authenticated local application;
            their hosted views are read-only help/imports. Real Orca qualification remains NOT RUN.
          </p>
        </header>
        <div className="module-grid">
          {modules.map(([title, text], i) => (
            <article className="module-card" key={title}>
              <span className="eyebrow">0{i + 1}</span>
              <h3>{title}</h3>
              <p>{text}</p>
              <a href={"/modules/" + ["hysteresis", "crossbar", "mnist", "circuits", "comparison", "eda", "docs"][i]}>Explore {title}</a>
            </article>
          ))}
        </div>
      </section>
      <section id="science" className="section">
        <p className="eyebrow">Accuracy first</p>
        <h2>Built to teach the physics without overstating the hardware.</h2>
        <p>
          Material presets and conductance levels are simulation defaults for
          visualization and education. Literature-backed assumptions are
          surfaced plainly.
        </p>
        <ul>
          <li>
            Published ferroelectric model families: Materlik, Park, Alessandri,
            and Guo.
          </li>
          <li>
            Crossbar non-idealities include IR drop, sneak-path current, and
            conductance drift.
          </li>
          <li>
            EDA outputs are for experimentation, validation, and reproducible
            teaching artifacts.
          </li>
        </ul>
        <p>
          Model background:{" "}
          <a href="https://doi.org/10.1063/1.4916229">Materlik et al. (2015)</a>
          . Read the{" "}
          <a href="https://github.com/TrebuchetDynamics/fecim-lattice-tools/blob/development/docs/research/honesty-audit.md">
            honesty audit
          </a>{" "}
          and{" "}
          <a href="https://github.com/TrebuchetDynamics/fecim-lattice-tools">
            source and documentation
          </a>
          .
        </p>
      </section>
      <section id="start" className="section">
        <p className="eyebrow">Local tools</p>
        <h2>Clone, test, and launch the simulator.</h2>
        <pre>
          <code>{`git clone https://github.com/TrebuchetDynamics/fecim-lattice-tools.git
cd fecim-lattice-tools
go test ./...
go run ./cmd/fecim-lattice-tools`}</code>
        </pre>
        <p>
          The existing command opens the legacy desktop application and needs
          its native build dependencies. This React application has hosted and
          local asset profiles. The additive authenticated local candidate is
          launched with fecim-lattice-tools serve --root DIRECTORY after building
          matching local assets. It is not the default desktop launch and is not
          human Orca-qualified. Hosted local-only actions show instructions and
          never probe localhost. No runtime Node is required for either distribution.
        </p>
      </section>
    </>
  );
}
