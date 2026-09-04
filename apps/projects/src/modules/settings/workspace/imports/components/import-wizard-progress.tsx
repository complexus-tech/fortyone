import { CheckIcon } from "icons";
import { cn } from "lib";
import { Box, Flex, Text } from "ui";

const STEPS = ["Upload", "Teams", "Members", "Review", "Import"] as const;
export const WizardProgress = ({ step }: { step: number }) => (
  <Box className="mt-5">
    <Flex align="center" justify="between">
      {STEPS.map((label, index) => {
        const stepNumber = index + 1;
        const active = stepNumber <= step;
        return (
          <Flex align="center" className="min-w-0" gap={2} key={label}>
            <Box
              className={cn(
                "bg-surface-muted text-text-muted flex h-6 w-6 shrink-0 items-center justify-center rounded-full font-medium",
                active && "bg-foreground text-background",
              )}
            >
              {stepNumber < step ? <CheckIcon className="h-3.5" /> : stepNumber}
            </Box>
            <Text
              className={cn(
                "hidden font-medium md:block",
                !active && "text-text-muted",
              )}
            >
              {label}
            </Text>
          </Flex>
        );
      })}
    </Flex>
    <Box className="bg-border mt-3 h-1.5 overflow-hidden rounded-full">
      <Box
        className="bg-foreground h-full rounded-full transition-[width] duration-300"
        style={{ width: `${(step / STEPS.length) * 100}%` }}
      />
    </Box>
  </Box>
);
