export type EstimateScheme = "points" | "tshirt";

export type EstimateDisplayMode = "compact" | "full";

export const ESTIMATE_VALUES = [1, 2, 3, 5, 8] as const;

export type EstimateValue = (typeof ESTIMATE_VALUES)[number];

export const isEstimateValue = (value: number): value is EstimateValue =>
  ESTIMATE_VALUES.some((candidate) => candidate === value);

export const DEFAULT_ESTIMATE_SCHEME = "tshirt" satisfies EstimateScheme;

const compactEstimateLabels: Record<EstimateScheme, Record<number, string>> = {
  points: {
    1: "1",
    2: "2",
    3: "3",
    5: "5",
    8: "8",
  },
  tshirt: {
    1: "XS",
    2: "S",
    3: "M",
    5: "L",
    8: "XL",
  },
};

const unitLabels: Partial<Record<EstimateScheme, string>> = {
  points: "point",
};

const normalizeScheme = (scheme?: string | null): EstimateScheme => {
  if (scheme === "points" || scheme === "tshirt") {
    return scheme;
  }

  return DEFAULT_ESTIMATE_SCHEME;
};

export const formatEstimate = (
  scheme: string | null | undefined,
  value: number | null | undefined,
  mode: EstimateDisplayMode = "compact",
) => {
  if (!value) {
    return mode === "full" ? "No complexity" : "Complexity";
  }

  const normalizedScheme = normalizeScheme(scheme);
  const compactLabel =
    compactEstimateLabels[normalizedScheme][value] ?? String(value);

  if (mode === "compact" || normalizedScheme === "tshirt") {
    return compactLabel;
  }

  const unit = unitLabels[normalizedScheme] ?? "";
  const isSingular = compactLabel === "1";
  return `${compactLabel} ${unit}${isSingular ? "" : "s"}`;
};

export const getEstimateOptions = (scheme: string) => {
  const normalizedScheme = normalizeScheme(scheme);

  return ESTIMATE_VALUES.map((value) => ({
    label: formatEstimate(normalizedScheme, value, "compact"),
    value,
  }));
};
