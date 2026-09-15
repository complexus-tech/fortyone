import { cva } from "cva";

export const buttonVariants = cva("items-center justify-center", {
  variants: {
    size: {
      sm: "min-h-[44px] px-[12px] py-[10px]",
      md: "min-h-[44px] px-[16px] py-[10px]",
      lg: "min-h-[52px] px-[20px] py-[14px]",
    },
    color: {
      primary: "bg-primary-action dark:bg-primary-action-dark",
      invert: "bg-background-inverse dark:bg-background-inverse-dark",
      tertiary: "bg-surface-muted dark:bg-surface-muted-dark",
    },
    rounded: {
      none: "rounded-none",
      sm: "rounded-[6px]",
      md: "rounded-[10px]",
      lg: "rounded-[12px]",
      xl: "rounded-[20px]",
      full: "rounded-full",
    },
    disabled: {
      true: "opacity-40",
      false: "",
    },
    loading: {
      true: "opacity-80",
      false: "",
    },
    fullWidth: {
      true: "w-full",
      false: "",
    },
  },
  defaultVariants: {
    size: "md",
    rounded: "md",
    color: "primary",
    fullWidth: true,
  },
});
