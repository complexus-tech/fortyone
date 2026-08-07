export const formatDocumentRelativeTime = (value: string) => {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";

  const difference = Math.max(0, Date.now() - date.getTime());
  const minute = 60 * 1000;
  const hour = 60 * minute;
  const day = 24 * hour;

  if (difference < minute) return "Just now";
  if (difference < hour) {
    const minutes = Math.floor(difference / minute);
    return `${minutes} ${minutes === 1 ? "minute" : "minutes"} ago`;
  }
  if (difference < day) {
    const hours = Math.floor(difference / hour);
    return `${hours} ${hours === 1 ? "hour" : "hours"} ago`;
  }
  if (difference < 2 * day) return "Yesterday";

  return date.toLocaleDateString(undefined, {
    day: "numeric",
    month: "short",
    year:
      date.getFullYear() === new Date().getFullYear() ? undefined : "numeric",
  });
};
