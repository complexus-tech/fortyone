import type { ReactNode } from "react";
import { cn } from "lib";
import { Flex, Text } from "ui";

type SignalBannerRowProps = {
  ariaLabel?: string;
  actions?: ReactNode;
  children: ReactNode;
  className?: string;
  icon: ReactNode;
  title?: string;
  variant?: "embedded" | "standalone";
};

export const SignalBannerRow = ({
  ariaLabel,
  actions,
  children,
  className,
  icon,
  title,
  variant = "standalone",
}: SignalBannerRowProps) => (
  <Flex
    align="center"
    aria-label={ariaLabel}
    className={cn(
      "min-w-0 px-4 py-3 backdrop-blur-md",
      variant === "embedded"
        ? "rounded-none border-0 bg-transparent"
        : "border-primary/20 bg-primary/5 rounded-xl border",
      className,
    )}
    justify="between"
    role={ariaLabel ? "status" : undefined}
  >
    <Flex align="center" className="min-w-0 flex-1">
      {icon}
      <Text
        as="span"
        className="ml-2 min-w-0 truncate"
        color="primary"
        fontWeight="medium"
        title={title}
      >
        {children}
      </Text>
    </Flex>
    {actions ? (
      <Flex align="center" className="ml-2 shrink-0" gap={1}>
        {actions}
      </Flex>
    ) : null}
  </Flex>
);
