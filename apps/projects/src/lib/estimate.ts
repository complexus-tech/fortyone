export type EstimateScheme = "points" | "hours" | "tshirt" | "ideal_days";

export type EstimateDisplayMode = "compact" | "full";

export const ESTIMATE_VALUES = [1, 2, 3, 5, 8] as const;

export type EstimateValue = (typeof ESTIMATE_VALUES)[number];

const compactEstimateLabels: Record<EstimateScheme, Record<number, string>> = {
  points: {
    1: "1",
    2: "2",
    3: "3",
    5: "5",
    8: "8",
  },
  hours: {
    1: "0.5",
    2: "1",
    3: "2",
    5: "4",
    8: "8",
  },
  tshirt: {
    1: "XS",
    2: "S",
    3: "M",
    5: "L",
    8: "XL",
  },
  ideal_days: {
    1: "0.5",
    2: "1",
    3: "2",
    5: "3",
    8: "5",
  },
};

const unitLabels: Partial<Record<EstimateScheme, string>> = {
  points: "point",
  hours: "hour",
  ideal_days: "day",
};

const normalizeScheme = (scheme?: string | null): EstimateScheme => {
  if (
    scheme === "points" ||
    scheme === "hours" ||
    scheme === "tshirt" ||
    scheme === "ideal_days"
  ) {
    return scheme;
  }

  return "points";
};

export const formatEstimate = (
  scheme: EstimateScheme | string | null | undefined,
  value: number | null | undefined,
  mode: EstimateDisplayMode = "compact",
) => {
  if (!value) {
    return mode === "full" ? "No estimate" : "Estimate";
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

export const getEstimateOptions = (scheme: EstimateScheme | string) => {
  const normalizedScheme = normalizeScheme(scheme);

  return ESTIMATE_VALUES.map((value) => ({
    label: formatEstimate(normalizedScheme, value, "compact"),
    value,
  }));
};
