import type { components } from "./generated";
export type Schema<K extends keyof components["schemas"]> = components["schemas"][K];
export type Configuration = Schema<"Configuration"> & { name?: string };
export type Job = Schema<"Job">;
export type Query = Schema<"ResultsQuery">;
export type Analysis = Schema<"Analysis">;
export type Problem = Schema<"Problem">;
export type Field = Schema<"ConfigurationField">;
export const initialQuery: Query = { page: 1, page_size: 32, filter: "all", sort: "run_id", descending: false, revision: "" };
export interface ResearchAdapter {
 roots(): Promise<{ id: string; name: string }[]>;
 projects(page?: number): Promise<Schema<"Projects">>;
 open(root: string, path: string): Promise<Configuration>;
 create(root: string, name: string): Promise<Configuration>;
 example(): Promise<Configuration>;
 import(documents: Record<string, string>): Promise<Configuration>;
 configuration(id: string): Promise<Configuration>;
 save(id: string, revision: string, document: string, changes: Schema<"FieldChange">[]): Promise<Configuration>;
 validate(id: string): Promise<Schema<"Validation">>;
 close(id: string): Promise<void>;
 start(id: string, revision: string, workers: number, key: string): Promise<Job>;
 state(execution: string): Promise<Job>;
 cancel(execution: string): Promise<Job>;
 active(id: string): Promise<Job[]>;
 analysis(id: string, query: Query, compare?: string[]): Promise<Analysis>;
 inspect(id: string, run: string): Promise<string>;
 verify(id: string): Promise<Analysis>;
 reports(id: string): Promise<Schema<"Artifacts">>;
 export(id: string, run: string): Promise<Schema<"Artifacts">>;
 modules(): Promise<Schema<"ModuleDescriptor">[]>;
}
export class RequestError extends Error {
 constructor(readonly problem: Problem) { super(problem.detail); }
}
