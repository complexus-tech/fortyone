export const formatDocumentRelativeTime = (value: string) => {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";

  const difference = Math.max(0, Date.now() - date.getTime());
  const minute = 60 * 1000;
  const hour = 60 * minute;
  const day = 24 * hour;
  const week = 7 * day;

  if (difference < minute) return "Just now";
  if (difference < hour) {
    const minutes = Math.floor(difference / minute);
    return `${minutes}m`;
  }
  if (difference < day) {
    const hours = Math.floor(difference / hour);
    return `${hours}h`;
  }
  if (difference < week) {
    const days = Math.floor(difference / day);
    return `${days}d`;
  }

  return date.toLocaleDateString(undefined, {
    day: "numeric",
    month: "short",
    year:
      date.getFullYear() === new Date().getFullYear() ? undefined : "numeric",
  });
};
