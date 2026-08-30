import { Text } from "ui";
import {
  addDays,
  differenceInCalendarDays,
  differenceInDays,
  format,
} from "date-fns";

export const getDueDateMessage = (
  date: Date,
  storyTerm: string,
  now: Date | null = new Date(),
) => {
  if (!now) {
    return <Text fontSize="md">Due on {format(date, "MMM d, yyyy")}</Text>;
  }

  const daysUntilDue = differenceInDays(date, now);
  const isTomorrow = differenceInCalendarDays(date, now) === 1;

  if (date < now) {
    const daysOverdue = differenceInDays(now, date);
    if (daysOverdue === 0) {
      return (
        <>
          <Text fontSize="md">The {storyTerm} is due today</Text>
          <Text color="muted" fontSize="md">
            Zero days overdue
          </Text>
        </>
      );
    }
    if (daysOverdue === 1) {
      return (
        <>
          <Text fontSize="md">This was due yesterday</Text>
          <Text color="muted" fontSize="md">
            One day overdue
          </Text>
        </>
      );
    }
    return (
      <>
        <Text fontSize="md">This was due on {format(date, "MMM d, yyyy")}</Text>
        <Text color="muted" fontSize="md">
          {daysOverdue} days overdue
        </Text>
      </>
    );
  }
  if (date <= addDays(now, 7) && date >= now) {
    return (
      <>
        <Text fontSize="md">Due on {format(date, "MMM d, yyyy")}</Text>
        <Text color="muted" fontSize="md">
          {isTomorrow ? "Due tomorrow" : <>Due in {daysUntilDue + 1} days</>}
        </Text>
      </>
    );
  }
  return (
    <>
      <Text fontSize="md">Due on {format(date, "MMM d, yyyy")}</Text>
      <Text color="muted" fontSize="md">
        {isTomorrow ? "Tomorrow" : <>Due in {daysUntilDue + 1} days</>}
      </Text>
    </>
  );
};
