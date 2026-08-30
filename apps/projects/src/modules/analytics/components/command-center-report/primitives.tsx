import type { ReactNode } from "react";
import { Box, Flex, Skeleton, Text, Wrapper } from "ui";

type MetricCardProps = {
  accent?: string | null;
  description: string;
  label: string;
  value: string;
};

export const ReportCard = ({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) => {
  return (
    <Wrapper className={["h-full", className].filter(Boolean).join(" ")}>
      {children}
    </Wrapper>
  );
};

export const SectionTitle = ({
  children,
  description,
}: {
  children: ReactNode;
  description?: string;
}) => {
  return (
    <Box>
      <Text className="text-lg font-semibold">{children}</Text>
      {description ? (
        <Text className="text-foreground/70 mt-1 text-[0.92rem] leading-5">
          {description}
        </Text>
      ) : null}
    </Box>
  );
};

export const MetricCard = ({
  accent,
  description,
  label,
  value,
}: MetricCardProps) => {
  return (
    <Wrapper className="px-3 py-3 md:px-5 md:py-4">
      <Flex align="center" className="gap-4" justify="between">
        <Text className="text-2xl antialiased" fontWeight="semibold">
          {value}
        </Text>
        {accent ? (
          <Text className="text-success shrink-0 text-base font-medium">
            {accent}
          </Text>
        ) : null}
      </Flex>
      <Text className="mt-2 opacity-80" color="muted">
        {label}
      </Text>
      <Text
        className="text-foreground/70 mt-1 truncate text-[0.9rem] leading-5"
        title={description}
      >
        {description}
      </Text>
    </Wrapper>
  );
};

export const MiniMetric = ({
  description,
  label,
  value,
}: {
  description: string;
  label: string;
  value: string;
}) => {
  return (
    <Wrapper className="py-3">
      <Text className="text-[1.45rem] leading-none font-semibold">{value}</Text>
      <Text className="mt-2 font-medium">{label}</Text>
      <Text className="text-foreground/70 mt-1 text-[0.9rem] leading-5">
        {description}
      </Text>
    </Wrapper>
  );
};

export const EmptyState = ({ children }: { children: ReactNode }) => {
  return (
    <Flex
      align="center"
      className="bg-surface-muted/40 min-h-28 rounded-lg px-4 py-6 text-center"
      justify="center"
    >
      <Text color="muted">{children}</Text>
    </Flex>
  );
};

export const CommandCenterSkeleton = () => {
  return (
    <Box className="pt-3 pb-5">
      <Flex className="mb-6 flex-col gap-3 @3xl:flex-row @3xl:items-end @3xl:justify-between">
        <Box>
          <Skeleton className="mb-3 h-8 w-72" />
          <Skeleton className="h-5 w-full max-w-xl" />
        </Box>
        <Skeleton className="h-10 w-80" />
      </Flex>
      <Box className="mb-6 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {Array.from({ length: 8 }).map((_, index) => (
          <Skeleton className="h-32" key={index} />
        ))}
      </Box>
      <Box className="grid gap-5 @6xl:grid-cols-2">
        <Skeleton className="h-[28rem]" />
        <Skeleton className="h-[28rem]" />
      </Box>
    </Box>
  );
};
