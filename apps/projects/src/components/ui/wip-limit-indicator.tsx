import { WarningIcon } from "icons";
import { Text, Tooltip } from "ui";

export const isWipLimitExceeded = (
  activeCount: number,
  limit?: number | null,
) => Boolean(limit && activeCount > limit);

export const WipLimitIndicator = ({
  activeCount,
  limit,
  taskTerm = activeCount === 1 ? "task" : "tasks",
  pluralTaskTerm = "tasks",
}: {
  activeCount: number;
  limit?: number | null;
  taskTerm?: string;
  pluralTaskTerm?: string;
}) => {
  if (!limit) return null;

  const isOverLimit = isWipLimitExceeded(activeCount, limit);
  let capacityDescription = "";
  if (isOverLimit) {
    capacityDescription = `, over limit by ${activeCount - limit}`;
  } else if (activeCount === limit) {
    capacityDescription = ", at capacity";
  }

  return (
    <Tooltip
      title={
        <>
          <span className="block">
            {activeCount} {taskTerm} in this status · limit {limit}.
            {isOverLimit
              ? ` Reduce by ${activeCount - limit} to get within the limit.`
              : ""}
          </span>
          <span className="block">
            Includes {pluralTaskTerm} hidden by filters.
          </span>
        </>
      }
    >
      <Text
        as="span"
        className="focus-visible:ring-ring inline-flex shrink-0 items-center gap-1 rounded-sm whitespace-nowrap focus-visible:ring-2 focus-visible:outline-none"
        color={isOverLimit ? "warning" : "muted"}
        fontWeight={isOverLimit ? "medium" : undefined}
        tabIndex={0}
      >
        {isOverLimit ? (
          <WarningIcon
            aria-hidden
            className="text-warning h-4 w-auto shrink-0"
          />
        ) : null}
        <span aria-hidden>
          {activeCount}/{limit}
        </span>
        <span className="sr-only">
          {activeCount} {taskTerm}, limit {limit}
          {capacityDescription}
        </span>
      </Text>
    </Tooltip>
  );
};
