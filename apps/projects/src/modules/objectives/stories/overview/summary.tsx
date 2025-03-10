import { CalendarIcon, HealthIcon, OKRIcon } from "icons";
import { Box, Text, Wrapper, ProgressBar, Flex, DatePicker } from "ui";
import { useParams } from "next/navigation";
import { cn } from "lib";
import { differenceInDays, format } from "date-fns";
import { useSession } from "next-auth/react";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { useObjective } from "../../hooks/use-objective";
import { useKeyResults } from "../../hooks/use-key-results";
import { useUpdateObjectiveMutation } from "../../hooks";
import type { KeyResult } from "../../types";

const getProgress = (keyResult: KeyResult) => {
  if (keyResult.measurementType === "boolean") {
    return keyResult.currentValue === 1 ? 100 : 0;
  }
  if (keyResult.measurementType === "percentage") {
    return keyResult.currentValue;
  }
  // Calculate progress relative to start value
  const totalChange = keyResult.targetValue - keyResult.startValue;
  const actualChange = keyResult.currentValue - keyResult.startValue;
  return Math.round((actualChange / totalChange) * 100);
};

export const Summary = () => {
  const { data: session } = useSession();
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const updateObjective = useUpdateObjectiveMutation();
  const { data: objective } = useObjective(objectiveId);
  const { data: keyResults = [] } = useKeyResults(objectiveId);
  const { isAdminOrOwner } = useIsAdminOrOwner(objective?.createdBy);
  const canUpdate = isAdminOrOwner || session?.user?.id === objective?.leadUser;

  const progress =
    Math.round(
      ((objective?.stats.completed || 0) / (objective?.stats.total || 1)) * 100,
    ) || 0;

  const getTargetDateMessage = () => {
    const targetDate = new Date(objective?.endDate || "");
    const today = new Date();
    const days = differenceInDays(targetDate, today);

    if (days <= 0) {
      return "Target date has passed";
    }

    if (days === 1) {
      return "1 day left";
    }

    return `${days} days left`;
  };

  const handleSetTargetDate = (date: string) => {
    updateObjective.mutate({
      objectiveId,
      data: { endDate: date },
    });
  };

  const keyResultProgress = Math.round(
    keyResults.reduce((acc, keyResult) => {
      return acc + getProgress(keyResult);
    }, 0) / keyResults.length,
  );

  return (
    <Box
      className={cn("mt-3 grid grid-cols-2 gap-4", {
        "grid-cols-3": keyResults.length > 0,
      })}
    >
      <Wrapper className="px-5">
        <Text
          className="mb-2 flex items-center gap-1.5 antialiased"
          fontSize="lg"
          fontWeight="semibold"
        >
          <HealthIcon />
          Progress
        </Text>
        <Text fontSize="2xl">
          {progress}%{" "}
          <Text as="span" color="muted" fontSize="md">
            completed
          </Text>
        </Text>
        <ProgressBar className="mt-2.5" progress={progress} />
      </Wrapper>
      {keyResults.length > 0 && (
        <Wrapper className="px-5">
          <Text
            className="mb-2 flex items-center gap-1.5 antialiased"
            fontSize="lg"
            fontWeight="semibold"
          >
            <OKRIcon />
            Key Result Progress
          </Text>
          <Text fontSize="2xl">
            {keyResultProgress}%{" "}
            <Text as="span" color="muted" fontSize="md">
              completed
            </Text>
          </Text>
          <ProgressBar className="mt-2.5" progress={keyResultProgress} />
        </Wrapper>
      )}
      <Wrapper className="px-5">
        <Text
          className="mb-2 flex items-center gap-1 antialiased"
          fontSize="lg"
          fontWeight="semibold"
        >
          <CalendarIcon />
          Target date
        </Text>
        {objective?.endDate ? (
          <>
            <Text fontSize="2xl">
              {format(new Date(objective.endDate), "MMM d")}
              <Text as="span" color="muted" fontSize="md">
                , {new Date(objective.endDate).getFullYear()}
              </Text>
            </Text>
            <Text as="span" className="text-[0.95rem]" color="muted">
              {getTargetDateMessage()}
            </Text>
          </>
        ) : (
          <Flex align="center" gap={1}>
            <CalendarIcon className="relative -left-1 h-10 opacity-30" />
            <Box>
              <Text color="muted">No target date</Text>
              <DatePicker>
                <DatePicker.Trigger>
                  <button
                    className="text-primary disabled:cursor-not-allowed disabled:opacity-50"
                    disabled={!canUpdate}
                    type="button"
                  >
                    Set target date
                  </button>
                </DatePicker.Trigger>
                <DatePicker.Calendar
                  onDayClick={(day) => {
                    if (canUpdate) {
                      handleSetTargetDate(day.toISOString());
                    }
                  }}
                  selected={
                    objective?.startDate
                      ? new Date(objective.startDate)
                      : undefined
                  }
                />
              </DatePicker>
            </Box>
          </Flex>
        )}
      </Wrapper>
    </Box>
  );
};
