import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { cn } from "lib";

type PricingIconProps = ComponentPropsWithoutRef<"svg">;

const PricingIcon = ({
  children,
  className,
  ...props
}: PricingIconProps & { children: ReactNode }) => (
  <svg
    {...props}
    className={cn("size-5", className)}
    fill="none"
    stroke="currentColor"
    strokeLinecap="round"
    strokeLinejoin="round"
    strokeWidth="1.8"
    viewBox="0 0 24 24"
    xmlns="http://www.w3.org/2000/svg"
  >
    {children}
  </svg>
);

export const PricingArrowIcon = (props: PricingIconProps) => (
  <PricingIcon {...props}>
    <path d="M18.5 12H5" />
    <path d="M13 18s6-4.419 6-6-6-6-6-6" />
  </PricingIcon>
);

export const PricingAiIcon = (props: PricingIconProps) => (
  <PricingIcon {...props}>
    <path d="M12 3c.55 4.85 3.15 7.45 8 8-4.85.55-7.45 3.15-8 8-.55-4.85-3.15-7.45-8-8 4.85-.55 7.45-3.15 8-8Z" />
    <path d="M19 3v4" />
    <path d="M17 5h4" />
  </PricingIcon>
);

export const PricingFeedbackIcon = (props: PricingIconProps) => (
  <PricingIcon {...props}>
    <path d="M5 4h14a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2h-8l-6 4v-4a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2Z" />
    <path d="M8 9h8" />
    <path d="M8 13h5" />
  </PricingIcon>
);

export const PricingIntegrationIcon = (props: PricingIconProps) => (
  <PricingIcon {...props}>
    <path d="M8.5 3v4.5" />
    <path d="M15.5 3v4.5" />
    <path d="M6 7.5h12v2A6 6 0 0 1 12 15.5v0A6 6 0 0 1 6 9.5v-2Z" />
    <path d="M12 15.5V21" />
    <path d="M9 21h6" />
  </PricingIcon>
);

export const PricingWorkIcon = (props: PricingIconProps) => (
  <PricingIcon {...props}>
    <circle cx="6" cy="6" r="2" />
    <circle cx="18" cy="6" r="2" />
    <circle cx="12" cy="18" r="2" />
    <path d="M8 6h8" />
    <path d="m7.2 7.8 3.6 8.4" />
    <path d="m16.8 7.8-3.6 8.4" />
  </PricingIcon>
);

export const PricingAccessIcon = (props: PricingIconProps) => (
  <PricingIcon {...props}>
    <path d="M12 3 19 6v5c0 4.5-2.8 7.8-7 10-4.2-2.2-7-5.5-7-10V6l7-3Z" />
    <path d="m9 12 2 2 4-4" />
  </PricingIcon>
);

export const PricingTeamIcon = (props: PricingIconProps) => (
  <PricingIcon {...props}>
    <circle cx="12" cy="7" r="3" />
    <circle cx="5" cy="11" r="2" />
    <circle cx="19" cy="11" r="2" />
    <path d="M7 20v-2c0-2.4 2.2-4 5-4s5 1.6 5 4v2" />
    <path d="M2.5 19v-1c0-1.7 1.2-2.8 3-3" />
    <path d="M21.5 19v-1c0-1.7-1.2-2.8-3-3" />
  </PricingIcon>
);

export const PricingPlanningIcon = (props: PricingIconProps) => (
  <PricingIcon {...props}>
    <circle cx="7" cy="7" r="4" />
    <circle cx="7" cy="7" r="1" />
    <path d="M11 7h4a3 3 0 0 1 3 3v6" />
    <path d="m15 14 3 3 3-3" />
    <path d="M4 16h7" />
    <path d="M4 20h4" />
  </PricingIcon>
);

export const PricingOrganizationIcon = (props: PricingIconProps) => (
  <PricingIcon {...props}>
    <rect height="5" rx="1.75" width="8" x="8" y="3" />
    <rect height="5" rx="1.75" width="7" x="3" y="16" />
    <rect height="5" rx="1.75" width="7" x="14" y="16" />
    <path d="M12 8v4" />
    <path d="M6.5 16v-2a2 2 0 0 1 2-2h7a2 2 0 0 1 2 2v2" />
  </PricingIcon>
);

export const PricingAdminIcon = (props: PricingIconProps) => (
  <PricingIcon {...props}>
    <circle cx="8" cy="12" r="4" />
    <path d="M12 12h8" />
    <path d="M17 12v3" />
    <path d="M20 12v2" />
  </PricingIcon>
);

export const PricingSupportIcon = (props: PricingIconProps) => (
  <PricingIcon {...props}>
    <path d="M4 5.5h10a3 3 0 0 1 3 3v4a3 3 0 0 1-3 3H9l-4.5 3v-3.6A3 3 0 0 1 2 12V8.5a3 3 0 0 1 2-3Z" />
    <path d="M17 9h1a4 4 0 0 1 4 4v1a4 4 0 0 1-3 3.9V21l-4.5-3H12" />
  </PricingIcon>
);

export const PricingDeploymentIcon = (props: PricingIconProps) => (
  <PricingIcon {...props}>
    <path d="m12 3 8 4.5-8 4.5-8-4.5L12 3Z" />
    <path d="m4 12 8 4.5 8-4.5" />
    <path d="m4 16.5 8 4.5 8-4.5" />
  </PricingIcon>
);

export const PricingScaleIcon = (props: PricingIconProps) => (
  <PricingIcon {...props}>
    <circle cx="12" cy="12" r="2.5" />
    <path d="m9 9-5-5" />
    <path d="M4 8V4h4" />
    <path d="m15 9 5-5" />
    <path d="M16 4h4v4" />
    <path d="m9 15-5 5" />
    <path d="M4 16v4h4" />
    <path d="m15 15 5 5" />
    <path d="M16 20h4v-4" />
  </PricingIcon>
);
