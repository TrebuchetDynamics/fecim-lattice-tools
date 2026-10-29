import type { Core } from "./types";
// Only coordinate normalization for a supplementary drawing; no scientific
// interpolation, metrics or filtering. All underlying rows are in Go pages.
export function Plot({ core }: { core: Core }) {
  const points = core.Plot || [];
  if (!points.length) return <p>No plot samples. Run or advance the loop.</p>;
  const es = points.map((p) => p.E),
    ps = points.map((p) => p.P);
  const minE = Math.min(...es),
    maxE = Math.max(...es),
    minP = Math.min(...ps),
    maxP = Math.max(...ps);
  const coords = points
    .map((p) =>
      [
        20 + ((p.E - minE) / (maxE - minE || 1)) * 560,
        280 - ((p.P - minP) / (maxP - minP || 1)) * 260,
      ].join(","),
    )
    .join(" ");
  return (
    <figure>
      <svg
        viewBox="0 0 600 300"
        role="img"
        aria-labelledby="plot-title plot-desc"
      >
        <title id="plot-title">
          Supplementary polarization–field trajectory
        </title>
        <desc id="plot-desc">
          Bounded Go plot samples, in history order. Read every retained value
          in the paged loop table. This plot has no unique controls.
        </desc>
        <polyline
          points={coords}
          fill="none"
          stroke="#175e4b"
          strokeWidth="2"
        />
      </svg>
      <figcaption>
        E horizontally, P vertically; canonical Go coordinates. Full history is
        available in the table, including points omitted from this bounded
        drawing.
      </figcaption>
    </figure>
  );
}
