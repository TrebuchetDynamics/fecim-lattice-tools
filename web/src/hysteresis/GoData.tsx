import type { Quantity } from "./types";

export function isQuantity(value: unknown): value is Quantity {
  return (
    !!value &&
    typeof value === "object" &&
    "Presentation" in value &&
    "SI" in value
  );
}
export function QuantityValue({ value }: { value: Quantity }) {
  return (
    <>
      <span aria-hidden="true">{value.Presentation.Display}</span>
      <span className="visually-hidden">{value.Presentation.Speakable}</span>
      <small>
        Canonical: {String(value.SI)} {value.Unit}
      </small>
      {value.Presentation.Reason && <small>{value.Presentation.Reason}</small>}
    </>
  );
}
// Faithfully expose nested Go metadata, including unavailable reasons and
// calibration/reference arrays. No JSON blobs or plot-only scientific values.
export function GoData({
  value,
  label = "Go record",
  values = {},
}: {
  value: unknown;
  label?: string;
  values?: Record<string, Quantity>;
}) {
  const rows: { path: string; value: unknown }[] = [];
  function visit(item: unknown, path: string) {
    if (values[path]) {
      rows.push({ path, value: values[path] });
      return;
    }
    if (item && typeof item === "object" && !isQuantity(item)) {
      const entries = Object.entries(item);
      if (!entries.length) rows.push({ path, value: "No entries" });
      for (const [key, child] of entries)
        visit(child, path ? path + " / " + key : key);
    } else rows.push({ path, value: item });
  }
  visit(value, "");
  return (
    <table className="data-tree">
      <caption>{label}</caption>
      <tbody>
        {rows.map((row, index) => (
          <tr key={index}>
            <th scope="row">{row.path || label}</th>
            <td>
              {isQuantity(row.value) ? (
                <QuantityValue value={row.value} />
              ) : row.value === undefined || row.value === null ? (
                "Not supplied"
              ) : (
                String(row.value)
              )}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
export function ResultsTable<R extends object>({
  id,
  page,
  keys,
  caption,
  captionRef,
  labels = {},
}: {
  id: string;
  page: { Rows: R[] };
  keys: (keyof R)[];
  caption: string;
  labels?: Record<string, string>;
  captionRef?: React.Ref<HTMLTableCaptionElement>;
}) {
  return (
    <div
      className="table-scroll"
      role="region"
      aria-label={
        id === "core-table" ? "Loop data table" : "Diagnostic data table"
      }
      tabIndex={0}
    >
      <table id={id}>
        <caption ref={captionRef} tabIndex={-1}>
          {caption}
        </caption>
        <thead>
          <tr>
            {keys.map((key) => (
              <th scope="col" key={String(key)}>
                {labels[String(key)] || String(key)}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {page.Rows.map((row, index) => (
            <tr key={index}>
              {keys.map((key) => {
                const value = row[key];
                return key === "Index" ? (
                  <th scope="row" key={String(key)}>
                    {String(value)}
                  </th>
                ) : isQuantity(value) ? (
                  <td
                    key={String(key)}
                    data-speech={value.Presentation.Speakable}
                    aria-label={value.Presentation.Speakable}
                  >
                    {value.Presentation.Display}
                    <span className="visually-hidden">
                      {" "}
                      {value.Presentation.Reason}
                    </span>
                  </td>
                ) : (
                  <td key={String(key)}>{String(value)}</td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
