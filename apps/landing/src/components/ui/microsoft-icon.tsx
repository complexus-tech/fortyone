import type { ComponentPropsWithoutRef } from "react";
import { cn } from "lib";

export const MicrosoftIcon = ({
  className,
  ...props
}: ComponentPropsWithoutRef<"svg">) => {
  return (
    <svg
      className={cn("h-6 w-auto", className)}
      fill="none"
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path d="M2 2h9.5v9.5H2z" fill="#F25022" />
      <path d="M12.5 2H22v9.5h-9.5z" fill="#7FBA00" />
      <path d="M2 12.5h9.5V22H2z" fill="#00A4EF" />
      <path d="M12.5 12.5H22V22h-9.5z" fill="#FFB900" />
    </svg>
  );
};
