import { cloneElement, useId } from "react";
import type { ReactElement, ReactNode } from "react";
export function Field({
  label,
  hint,
  children,
}: {
  label: ReactNode;
  hint?: string;
  children: ReactElement<{ id?: string; "aria-describedby"?: string }>;
}) {
  const id = useId();
  return (
    <div className="field">
      <label htmlFor={id}>{label}</label>
      {cloneElement(children, {
        id,
        "aria-describedby": hint ? id + "-help" : undefined,
      })}
      {hint && (
        <span className="hint" id={id + "-help"}>
          {hint}
        </span>
      )}
    </div>
  );
}
