import { lazy, Suspense } from "react";
import { createRoot } from "react-dom/client";
import { Home } from "./Home";
import "./style.css";
const Hysteresis = lazy(() => import("./hysteresis/Hysteresis"));
const routes = { "/": Home, "/hysteresis": Hysteresis };
const View =
  routes[(location.pathname.replace(/\/$/, "") as keyof typeof routes) || "/"];
createRoot(document.getElementById("root")!).render(
  <>
    <a className="skip-link" href="#main">
      Skip to content
    </a>
    <header className="nav">
      <a className="brand" href="/">
        <b className="brand-mark" aria-hidden="true">
          Fe
        </b>{" "}
        FeCIM Lattice Tools
      </a>
      <nav className="nav__links" aria-label="Primary navigation">
        <a href="/hysteresis">Hysteresis lab</a>
        <a href="/#modules">Modules</a>
        <a href="/#science">Science</a>
        <a href="/#start">Get started</a>
      </nav>
    </header>
    <main id="main">
      {View ? (
        <Suspense fallback={<p>Loading interface…</p>}>
          <View />
        </Suspense>
      ) : (
        <>
          <h1>Page not found</h1>
          <a href="/">Home</a>
        </>
      )}
    </main>
    <footer>
      FeCIM / Simulation only · MIT · Go scientific core
    </footer>
  </>,
);
