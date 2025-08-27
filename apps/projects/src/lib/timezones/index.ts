import tzs from "./timezones.json";

export type TimeZone = {
  label: string;
  tzCode: string;
  name: string;
  utc: string;
};

export const timezones = tzs as TimeZone[];
