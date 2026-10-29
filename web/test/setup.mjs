// Private test setup only. CI supplies RUNNER_TEMP; standalone Node/Playwright
// use the platform temp root. Autowork's explicit TMPDIR always takes priority.
import { tmpdir } from "node:os";
process.env.TMPDIR ||= process.env.RUNNER_TEMP || tmpdir();
