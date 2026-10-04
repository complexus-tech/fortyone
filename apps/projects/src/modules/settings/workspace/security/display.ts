const SECURITY_DATE_FORMAT = new Intl.DateTimeFormat("en-GB", {
  day: "2-digit",
  month: "short",
  year: "numeric",
  timeZone: "UTC",
});
const SECURITY_TIME_FORMAT = new Intl.DateTimeFormat("en-GB", {
  hour: "2-digit",
  minute: "2-digit",
  hourCycle: "h23",
  timeZone: "UTC",
  timeZoneName: "short",
});

export const securityDateParts = (value: string) => {
  const date = new Date(value);
  return {
    date: SECURITY_DATE_FORMAT.format(date),
    time: SECURITY_TIME_FORMAT.format(date),
  };
};
export const securityDate = (value: string) => {
  const parts = securityDateParts(value);
  return `${parts.date} at ${parts.time}`;
};

export const securityLabel = (value: string) => {
  const label = value
    .replace(/[._]/g, " ")
    .replace(/\b(?:api|scim|sso|oauth)\b/gi, (word) => word.toUpperCase());
  return label.charAt(0).toUpperCase() + label.slice(1);
};
