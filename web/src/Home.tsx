import "./Home.css";

const labGroups = [
  {
    title: "Model devices and circuits",
    modules: [
      ["hysteresis", "Hysteresis Lab", "Explore ferroelectric switching with HZO, BTO, and PZT presets."],
      ["circuits", "Peripheral Circuits", "Inspect DAC, ADC, TIA, read paths, and ISPP write control."],
    ],
  },
  {
    title: "Study arrays and systems",
    modules: [
      ["crossbar", "Crossbar Array", "Simulate matrix-vector multiply with quantized conductance, IR drop, sneak paths, and drift."],
      ["comparison", "Technology Comparison", "Compare assumptions, operating points, and metrics across memory technologies."],
      ["mnist", "MNIST Pipeline", "Inspect the CIM inference path from input quantization to accuracy reporting."],
    ],
  },
  {
    title: "Export designs and read references",
    modules: [
      ["eda", "EDA Export", "Explore SPICE, Verilog, Liberty, DEF, and LEF artifacts for education-stage experiments."],
      ["docs", "Integrated Docs", "Find physics explanations, references, module guides, and honesty notes."],
    ],
  },
] as const;

const local = __FECIM_PROFILE__ === "local";

export function Home() {
  return (
    <>
      <header className="home-intro">
        <p className="eyebrow">FeCIM Lattice Tools · Ferroelectric compute-in-memory</p>
        <h1>Research workbench</h1>
      </header>
      <div className="home-overview">
        <nav className="home-panel" aria-labelledby="home-study-title">
          <h2 id="home-study-title">Start a study</h2>
          <p>
            {local
              ? "Local application: open a project from your granted directory in an authenticated session."
              : "Hosted browser: open a small example and run a bounded study in memory."}
          </p>
          <div className="actions">
            <a className="button primary" href="/projects">Open Projects</a>
            <a className="button" href="/#start">Run locally</a>
          </div>
          <p className="honesty">
            Education-phase simulation — material presets and 30 default
            conductance levels are educational assumptions, not validated device
            measurements.
          </p>
        </nav>
        <nav className="home-panel" aria-labelledby="home-labs-title">
          {/* Keep the fragment focus on the introduction, not the full list. */}
          <header id="modules">
            <h2 id="home-labs-title">Explore a lab</h2>
            <p>
              MNIST inference and EDA generation require the authenticated local application;
              their hosted views provide read-only help and imports. Real Orca qualification remains NOT RUN.
            </p>
          </header>
          <div className="module-grid home-lab-groups">
            {labGroups.map(({ title, modules }) => (
              <div className="home-lab-group" key={title}>
                <h3>{title}</h3>
                <ul>
                  {modules.map(([id, name, description]) => (
                    <li className="module-card" key={id}>
                      <a href={`/modules/${id}`}>Explore {name}</a>
                      <p>{description}</p>
                      {(id === "mnist" || id === "eda") && (
                        <p className="home-capability">
                          Hosted: help and imports only. Local: {id === "mnist" ? "run inference" : "generate artifacts"} with authentication.
                        </p>
                      )}
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </nav>
      </div>
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
