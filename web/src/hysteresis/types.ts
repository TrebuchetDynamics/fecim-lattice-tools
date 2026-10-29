export type JSONValue =
  | null
  | boolean
  | string
  | number
  | JSONValue[]
  | { [key: string]: JSONValue };
export interface Identity {
  protocol: string;
  model: string;
  build: string;
  go: string;
}
export interface Reply<T> {
  identity: Identity;
  session: string;
  request: number;
  revision: number;
  operation: string;
  result?: T;
  error?: string;
  committed: boolean;
}
export interface Presentation {
  Display: string;
  Speakable: string;
  Availability: string;
  Reason: string;
}
export interface Quantity {
  SI: number;
  Unit: string;
  Presentation: Presentation;
}
export interface Config {
  MaterialID: string;
  Model: string;
  Mode: string;
  Frequency: number;
  TimeScale: number;
  Temperature: number;
  Stress: number;
  Levels: number;
  RangeFraction: number;
  HistoryLimit: number;
  Seed: number;
}
export interface Control {
  Enabled: boolean;
  Reason: string;
  Unit: string;
  Min: number;
  Max: number;
  Step: number;
  MinExclusive: boolean;
  SliderMin: number;
  SliderMax: number;
}
export interface Material {
  ID: string;
  Parameters: { Name: string; [key: string]: JSONValue };
  Evidence: string;
  DefaultLevels: number;
  DefaultRangeFraction: number;
}
export interface Core {
  ScientificValues: Record<string, Quantity>;
  Revision: number;
  ConfigRevision: number;
  Config: Config;
  State: string;
  View: string;
  Material: Material;
  Catalogue: Material[];
  Models: string[];
  Modes: string[];
  Evidence: string;
  Assumptions: string[];
  Controls: Record<string, Control>;
  E: Quantity;
  P: Quantity;
  Time: number;
  Level: number;
  Count: number;
  Dropped: number;
  SolverSteps: number;
  ViewCount: number;
  Plot: { Index: number; Time: number; E: number; P: number }[];
}
export interface CoreRow {
  TimeValue: Quantity;
  Index: number;
  Time: number;
  Branch: string;
  E: Quantity;
  P: Quantity;
}
export interface Page<R> {
  Revision: number;
  Offset: number;
  Total: number;
  Position?: string;
  Rows: R[];
}
export interface ProgramConfig {
  Mode: string;
  Method: string;
  Domains: number;
  WorkLimit: number;
  PulseLimit: number;
  CalibrationWorkLimit: number;
  GuardFraction: number;
}
export interface Program {
  Core: Core;
  Config: ProgramConfig;
  Status: string;
  Reason: string;
  Phase: string;
  Target: number;
  QueuedTarget: number;
  ReadLevel: number;
  Work: number;
  OwnsDrive: boolean;
  Controller: JSONValue;
  Calibration: { Status: string; Reason: string; [key: string]: JSONValue };
}
export interface FORC {
  EmaxVm: number;
  Reversals: number;
  Resolution: number;
}
export interface Progress {
  Revision: number;
  State: string;
  Completed: number;
  Total: number;
}
export interface AnalysisResult {
  AppliedParameters: Record<string, JSONValue>;
  ScientificValues: Record<string, Quantity>;
  Revision: number;
  Kind: string;
  Context: { Revision: number; ConfigRevision: number; Config: Config };
  Diagnostic: JSONValue;
  Evidence: string;
  ExperimentalValidation: boolean;
  Notices: string[];
  ReferenceContext: JSONValue;
  References: JSONValue;
  PUND: null | { Charges: JSONValue };
  FORC: JSONValue;
}
export interface AnalysisRow {
  TimeValue: Quantity;
  CurrentValue: Quantity;
  DensityValue: Quantity;
  Index: number;
  Pulse: string;
  Branch: string;
  TimeS: number;
  CurrentA: number;
  Density: number;
  E: Quantity;
  P: Quantity;
  Ea: Quantity;
  Eb: Quantity;
  Reversal: Quantity;
}
export interface Budget {
  Steps: number;
  DT: number;
}
export interface Catalogue {
  Identity: Identity;
  Program: ProgramConfig;
  ProgramModes: string[];
  Methods: string[];
  Diagnostics: string[];
  Tables: Record<string, string[]>;
  FORC: FORC;
  Views: string[];
  LoopBudget: Budget;
  ProgramBudget: Budget;
  AnalysisBudget: { Steps: number };
  PageSize: number;
  ReferenceBytes: number;
  Target: number;
  PlaybackCadenceMS: number;
  CoreTimeLabel: string;
  AnalysisColumns: Record<string, { Key: keyof AnalysisRow; Label: string }[]>;
}
export interface Manifest {
  API?: string;
  UI: string;
  Catalogue: Catalogue;
  identity: Identity;
  assets: Record<string, string>;
}
export interface SessionInfo {
  Core: Core;
  Operations: { name: string; mutates: boolean }[];
  MaxEnvelopeBytes: number;
  MaxAdvanceSteps: number;
}
export interface Results {
  core?: Core;
  corePage?: Page<CoreRow>;
  program?: Program;
  progress?: Progress;
  analysis?: AnalysisResult;
  analysisPage?: Page<AnalysisRow>;
  table?: string;
  revision: number;
  catalogue?: Catalogue;
  active?: "core" | "program" | "analysis";
  status: string;
  error?: string;
  failed: boolean;
}
declare global {
  const __FECIM_PROFILE__: "hosted" | "local";
  const __FECIM_UI__: string;
  const __FECIM_API__: string;
  var HysteresisWorkerClient: {
    new (options: Record<string, unknown>): {
      ready: Promise<void>;
      revision: number;
      state: string;
      request(op: string, input?: unknown): Promise<string>;
      terminate(state?: string): void;
    };
  };
}
