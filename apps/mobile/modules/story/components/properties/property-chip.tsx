import type { BadgeProps } from "@/components/ui/badge";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils/classnames";

export const PropertyChip = ({ className, ...props }: BadgeProps) => (
  <Badge
    {...props}
    rounded="full"
    color="tertiary"
    className={cn(
      "min-w-0 max-w-full shrink border-0 bg-accent px-[8px] py-[4px] dark:bg-accent-dark",
      className,
    )}
  />
);
