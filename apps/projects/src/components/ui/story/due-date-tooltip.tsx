import { Text } from "ui";
import { format, addDays, differenceInDays, isTomorrow } from "date-fns";

export const getDueDateMessage = (date: Date) => {
  if (date < new Date()) {
    const daysOverdue = differenceInDays(new Date(), date);
    if (daysOverdue === 0) {
      return (
        <>
          <Text fontSize="md">The story is due today</Text>
          <Text color="muted" fontSize="md">
            Zero days overdue
          </Text>
        </>
      );
    }
    if (daysOverdue === 1) {
      return (
        <>
          <Text fontSize="md">This was due on yesterday</Text>
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
          {differenceInDays(new Date(), date)} days overdue
        </Text>
      </>
    );
  }
  if (date <= addDays(new Date(), 7) && date >= new Date()) {
    return (
      <>
        <Text fontSize="md">Due on {format(date, "MMM d, yyyy")}</Text>
        <Text color="muted" fontSize="md">
          {isTomorrow(date) ? (
            "Due tomorrow"
          ) : (
            <>Due in {differenceInDays(date, new Date()) + 1} days</>
          )}
        </Text>
      </>
    );
  }
  return (
    <>
      <Text fontSize="md">Due on {format(date, "MMM d, yyyy")}</Text>
      <Text color="muted" fontSize="md">
        {isTomorrow(date) ? (
          "Tomorrow"
        ) : (
          <>Due in {differenceInDays(date, new Date()) + 1} days</>
        )}
      </Text>
    </>
  );
};
