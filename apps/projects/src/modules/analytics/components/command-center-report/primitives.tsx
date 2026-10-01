import type { ReactNode } from "react";
import { Box, Flex, Skeleton, Text, Wrapper } from "ui";
import { cn } from "lib";

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
    <Wrapper
      className={cn(
        "h-full min-w-0 rounded-2xl p-5 shadow-none md:p-6",
        className,
      )}
    >
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
      <Text className="text-lg font-medium">{children}</Text>
      {description ? (
        <Text className="mt-1 leading-6" color="muted">
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
    <Wrapper className="h-full min-w-0 rounded-2xl p-5 shadow-none md:p-5">
      <Text color="muted" fontWeight="medium">
        {label}
      </Text>
      <Flex align="end" className="mt-3 gap-3" justify="between">
        <Text
          className="text-3xl tracking-tight tabular-nums"
          fontWeight="medium"
        >
          {value}
        </Text>
        {accent ? (
          <Text className="shrink-0 tabular-nums" color="muted">
            {accent}
          </Text>
        ) : null}
      </Flex>
      <Text className="mt-3 leading-6" color="muted">
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
    <Wrapper className="min-w-0 py-4 shadow-none">
      <Text className="text-2xl leading-none font-medium tabular-nums">
        {value}
      </Text>
      <Text className="mt-2 font-medium">{label}</Text>
      <Text className="mt-2 leading-6" color="muted">
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

export const ChartLegend = ({
  items,
}: {
  items: { label: string; color: string }[];
}) => (
  <Flex className="mt-4 gap-x-5 gap-y-2" wrap>
    {items.map((item) => (
      <Flex align="center" gap={2} key={item.label}>
        <Box
          aria-hidden
          className="h-0.5 w-4 rounded-full"
          style={{ backgroundColor: item.color }}
        />
        <Text color="muted">{item.label}</Text>
      </Flex>
    ))}
  </Flex>
);

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
        {Array.from({ length: 4 }).map((_, index) => (
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
