import type {
  CustomField,
  CustomFieldValue,
  CustomFieldValueMap,
} from "./types";

const DECIMAL_PATTERN = /^-?\d+(?:\.\d+)?$/;
const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

export const isCustomFieldValueSet = (value: string | null | undefined) =>
  value !== undefined && value !== null && value !== "";

export const getCustomFieldInputLabel = (field: CustomField) =>
  field.type === "money" && field.currency
    ? `${field.name} (${field.currency})`
    : field.name;

export const createStoryValueBatches = (storyIds: string[]) => {
  const ids = Array.from(
    new Set(
      storyIds.filter((id) =>
        /^[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}$/i.test(id),
      ),
    ),
  ).sort();
  const batches: string[][] = [];
  for (let offset = 0; offset < ids.length; offset += 100)
    batches.push(ids.slice(offset, offset + 100));
  return batches;
};

export const validateCustomFieldValue = (
  field: CustomField,
  raw: string | null,
): string | null => {
  const value = raw?.trim();
  if (!value) return null;
  if (
    (field.type === "number" || field.type === "money") &&
    !DECIMAL_PATTERN.test(value)
  ) {
    return "Enter a decimal number, such as 1250.50.";
  }
  if (field.type === "date") {
    const parsed = new Date(`${value}T00:00:00Z`);
    if (
      !DATE_PATTERN.test(value) ||
      Number.isNaN(parsed.getTime()) ||
      parsed.toISOString().slice(0, 10) !== value
    ) {
      return "Choose a valid date.";
    }
  }
  if (
    field.type === "select" &&
    !field.options.some((option) => option.id === value && !option.archivedAt)
  ) {
    return "Choose an active option.";
  }
  if (
    field.type === "person" &&
    !/^[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}$/i.test(value)
  ) {
    return "Choose a team member.";
  }
  if (field.type === "text" && value.length > 10_000)
    return "Keep text within 10,000 characters.";
  return null;
};

export const prepareCustomFieldValues = (
  fields: CustomField[],
  values: CustomFieldValueMap,
): CustomFieldValue[] => {
  const result: CustomFieldValue[] = [];
  for (const field of fields) {
    if (field.archivedAt || !(field.id in values)) continue;
    const value = values[field.id]?.trim() || null;
    // Old template options may be archived without invalidating the template itself.
    if (
      field.type === "select" &&
      value &&
      !field.options.some((option) => option.id === value && !option.archivedAt)
    )
      continue;
    const error = validateCustomFieldValue(field, value);
    if (error) throw new Error(`${field.name}: ${error}`);
    result.push({ fieldId: field.id, value });
  }
  return result;
};

// Preserve decimal precision: report amounts can exceed JavaScript's safe integer range.
export const formatExactDecimal = (
  value: string,
  currency?: string | null,
  locale = "en-US",
) => {
  if (!DECIMAL_PATTERN.test(value)) return value;
  const negative = value.startsWith("-");
  const [whole = "0", fraction = ""] = value.replace(/^-/, "").split(".");
  let integer = BigInt(whole);
  let decimals = fraction;
  if (currency) {
    const precision =
      new Intl.NumberFormat(locale, {
        style: "currency",
        currency,
      }).resolvedOptions().maximumFractionDigits ?? 2;
    decimals = fraction.slice(0, precision).padEnd(precision, "0");
    const scale = BigInt(10) ** BigInt(precision);
    let scaled = integer * scale + BigInt(decimals || "0");
    if (Number(fraction[precision] ?? "0") >= 5) scaled += BigInt(1);
    integer = scaled / scale;
    decimals = precision
      ? (scaled % scale).toString().padStart(precision, "0")
      : "";
  }
  const numberFormat = new Intl.NumberFormat(locale);
  const decimalSeparator =
    numberFormat.formatToParts(1.1).find((part) => part.type === "decimal")
      ?.value ?? ".";
  const formatted = `${negative ? "−" : ""}${numberFormat.format(integer)}${decimals ? `${decimalSeparator}${decimals}` : ""}`;
  return currency ? `${currency} ${formatted}` : formatted;
};

export const formatCustomFieldValue = (
  field: CustomField,
  value: string | null,
  people: { id: string; fullName: string }[] = [],
) => {
  if (value === null || value === "") return "Not set";
  if (field.type === "money" || field.type === "number")
    return formatExactDecimal(value, field.currency);
  if (field.type === "select")
    return (
      field.options.find((option) => option.id === value)?.name ??
      "Unavailable option"
    );
  if (field.type === "person")
    return (
      people.find((person) => person.id === value)?.fullName ??
      "Unavailable member"
    );
  if (field.type === "date" && DATE_PATTERN.test(value)) {
    const parsed = new Date(`${value}T00:00:00Z`);
    if (Number.isNaN(parsed.getTime())) return "Invalid date";
    return new Intl.DateTimeFormat("en-US", {
      dateStyle: "medium",
      timeZone: "UTC",
    }).format(parsed);
  }
  return value;
};
