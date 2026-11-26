import type { Core } from "./types";

// Presentation geometry only. Go supplies the bounded history samples and SI
// units; no interpolation, derived metric or replacement science is computed.
export function Plot({ core }: { core: Core }) {
  const points = core.Plot || [];
  const es = points.map((p) => p.E), ps = points.map((p) => p.P);
  const minE = Math.min(0, ...es), maxE = Math.max(0, ...es);
  const minP = Math.min(0, ...ps), maxP = Math.max(0, ...ps);
  const x = (e: number) => 76 + ((e - minE) / (maxE - minE || 1)) * 600;
  const y = (p: number) => 324 - ((p - minP) / (maxP - minP || 1)) * 292;
  const coords = points.map((p) => `${x(p.E)},${y(p.P)}`).join(" ");
  const last = points.at(-1);
  return (
    <figure className="loop-figure">
      <svg
        className="loop-plot"
        viewBox="0 0 720 390"
        role="img"
        aria-labelledby="plot-title"
        aria-describedby="plot-desc"
      >
        <title id="plot-title">Supplementary polarization–field trajectory</title>
        <desc id="plot-desc">
          Bounded Go plot samples in history order. Electric field horizontally,
          polarization vertically, in canonical SI units. Read every retained
          value in this history view in the paged loop table. This plot has no
          unique controls.
        </desc>
        <rect x="76" y="32" width="600" height="292" fill="none" stroke="#c6cec9" />
        <path d={`M76 ${y(0)}H676 M${x(0)} 32V324`} stroke="#a3afa9" strokeDasharray="4 4" fill="none" />
        {points.length > 1 && (
          <polyline points={coords} fill="none" stroke="#175e4b" strokeWidth="2.5" strokeLinejoin="round" />
        )}
        {last && <circle cx={x(last.E)} cy={y(last.P)} r="4" fill="#175e4b" />}
        <text x="376" y="370" textAnchor="middle" className="loop-axis-label">Electric field ({core.E.Unit})</text>
        <text transform="translate(24 178) rotate(-90)" textAnchor="middle" className="loop-axis-label">Polarization ({core.P.Unit})</text>
      </svg>
      <figcaption>
        {points.length
          ? `Plot shows ${points.length} of ${core.ViewCount} Go samples in this history view. The table includes every sample in the view.`
          : "No loop samples. Run or advance the loop to generate Go simulation samples."}{" "}
        Core revision {core.Revision}.
      </figcaption>
    </figure>
  );
}
