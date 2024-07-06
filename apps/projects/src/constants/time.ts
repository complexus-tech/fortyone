const second = 1000;
const minute = 60 * second;
const hour = 60 * minute;
const day = 24 * hour;
const week = 7 * day;
const month = 30 * day;
const year = 365 * day;

export const DURATION_FROM_MILLISECONDS = {
  SECOND: second,
  MINUTE: minute,
  HOUR: hour,
  DAY: day,
  WEEK: week,
  MONTH: month,
  YEAR: year,
};

export const DURATION_FROM_SECONDS = {
  SECOND: second / second,
  MINUTE: minute / second,
  HOUR: hour / second,
  DAY: day / second,
  WEEK: week / second,
  MONTH: month / second,
  YEAR: year / second,
};
