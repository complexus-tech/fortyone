import { cva } from "cva";

export const textVariants = cva("text-foreground dark:text-foreground-dark", {
  variants: {
    align: {
      left: "text-left",
      center: "text-center",
      right: "text-right",
    },
    color: {
      foreground: "text-foreground dark:text-foreground-dark",
      inverse: "text-foreground-inverse dark:text-foreground-inverse-dark",
      primary: "text-primary dark:text-primary",
      primaryForeground: "text-primary-foreground dark:text-primary-foreground",
      secondaryForeground:
        "text-secondary-foreground dark:text-secondary-foreground-dark",
      secondary: "text-text-secondary dark:text-text-secondary-dark",
      muted: "text-text-muted dark:text-text-muted-dark",
      danger: "text-danger dark:text-danger-text-dark",
      warning: "text-warning dark:text-warning",
      info: "text-info dark:text-info",
      success: "text-success dark:text-success",
      black: "text-black dark:text-black",
      white: "text-white dark:text-white",
    },
    fontSize: {
      xs: "text-[13px] leading-[18px]",
      sm: "text-[15px] leading-[20px]",
      md: "text-[17px] leading-[25px]",
      lg: "text-[17px] leading-[24px]",
      xl: "text-[20px] leading-[26px]",
      "2xl": "text-[26px] leading-[32px]",
      "3xl": "text-[32px] leading-[38px]",
      "4xl": "text-[40px] leading-[46px]",
    },
    fontWeight: {
      light: "font-light",
      normal: "font-normal",
      medium: "font-medium",
      semibold: "font-semibold",
      bold: "font-bold",
    },
    transform: {
      uppercase: "uppercase",
      lowercase: "lowercase",
      capitalize: "capitalize",
      none: "normal-case",
    },
    fontStyle: {
      italic: "italic",
      normal: "not-italic",
    },
    decoration: {
      underline: "underline",
      lineThrough: "line-through",
      none: "no-underline",
    },
  },
  defaultVariants: {
    align: "left",
    fontSize: "md",
    fontWeight: "medium",
  },
});
