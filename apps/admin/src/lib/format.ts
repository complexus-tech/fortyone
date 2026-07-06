const dateFormatter = new Intl.DateTimeFormat("en", {
  month: "short",
  day: "numeric",
  year: "numeric",
});

const dateTimeFormatter = new Intl.DateTimeFormat("en", {
  month: "short",
  day: "numeric",
  year: "numeric",
  hour: "numeric",
  minute: "2-digit",
});

const numberFormatter = new Intl.NumberFormat("en");

export const formatCount = (value: number) => numberFormatter.format(value);

export const formatDate = (value: string | null | undefined) => {
  if (!value) {
    return "Not set";
  }
  return dateFormatter.format(new Date(value));
};

export const formatDateTime = (value: string | null | undefined) => {
  if (!value) {
    return "Not available";
  }
  return dateTimeFormatter.format(new Date(value));
};

export const daysFromNow = (value: string | null | undefined) => {
  if (!value) {
    return null;
  }

  const diff = new Date(value).getTime() - Date.now();
  return Math.ceil(diff / (1000 * 60 * 60 * 24));
};

export const formatTrialState = (value: string | null | undefined) => {
  const days = daysFromNow(value);

  if (days === null) {
    return "No trial";
  }
  if (days < 0) {
    return `${Math.abs(days)} days expired`;
  }
  if (days === 0) {
    return "Ends today";
  }
  if (days === 1) {
    return "1 day left";
  }
  return `${days} days left`;
};

export const humanizeKey = (value: string) =>
  value.replaceAll(".", " ").replaceAll("_", " ").replace(/\s+/g, " ").trim();

export const formatValue = (value: unknown) => {
  if (value === null || value === undefined || value === "") {
    return "None";
  }

  if (typeof value === "string") {
    return value;
  }

  if (typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }

  return JSON.stringify(value);
};
