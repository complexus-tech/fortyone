import { TextInput, TextInputProps, View } from "react-native";
import { VariantProps, cva } from "cva";
import { Text } from "./text";
import { cn } from "@/lib/utils";

const inputVariants = cva("border rounded-md bg-white px-4 py-3 text-base", {
  variants: {
    size: {
      sm: "px-3 py-2 text-sm min-h-[36px]",
      md: "px-4 py-3 text-base min-h-[44px]",
      lg: "px-5 py-4 text-lg min-h-[52px]",
    },
    variant: {
      default: "border-gray-200 bg-white",
      error: "border-danger bg-white",
      success: "border-success bg-white",
      disabled: "border-gray-200 bg-gray-50",
    },
    rounded: {
      none: "rounded-none",
      sm: "rounded-sm",
      md: "rounded-md",
      lg: "rounded-lg",
      xl: "rounded-xl",
      full: "rounded-full",
    },
  },
  defaultVariants: {
    size: "md",
    variant: "default",
    rounded: "md",
  },
});

export interface InputProps
  extends TextInputProps,
    VariantProps<typeof inputVariants> {
  label?: string;
  helpText?: string;
  errorText?: string;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  required?: boolean;
  containerClassName?: string;
  labelClassName?: string;
  helpTextClassName?: string;
}

export const Input = ({
  label,
  helpText,
  errorText,
  leftIcon,
  rightIcon,
  required,
  className,
  containerClassName,
  labelClassName,
  helpTextClassName,
  size,
  variant,
  rounded,
  editable = true,
  ...props
}: InputProps) => {
  const hasError = !!errorText;
  const inputVariant = hasError ? "error" : variant;
  const isDisabled = !editable;

  const inputClasses = inputVariants({
    size,
    variant: isDisabled ? "disabled" : inputVariant,
    rounded,
  });

  return (
    <View className={cn(containerClassName)}>
      {label && (
        <Text
          color="black"
          fontSize="sm"
          fontWeight="medium"
          className={cn("mb-2", labelClassName)}
        >
          {label}
          {required && <Text color="danger"> *</Text>}
        </Text>
      )}

      <View className="relative">
        {leftIcon && (
          <View className="absolute left-3 top-1/2 -translate-y-1/2 z-10">
            {leftIcon}
          </View>
        )}

        <TextInput
          className={cn(
            inputClasses,
            leftIcon ? "pl-10" : "",
            rightIcon ? "pr-10" : "",
            className
          )}
          editable={editable}
          placeholderTextColor="#6B665C"
          {...props}
        />

        {rightIcon && (
          <View className="absolute right-3 top-1/2 -translate-y-1/2 z-10">
            {rightIcon}
          </View>
        )}
      </View>

      {(helpText || errorText) && (
        <Text
          color={hasError ? "danger" : "muted"}
          fontSize="xs"
          className={cn("mt-1", helpTextClassName)}
        >
          {errorText || helpText}
        </Text>
      )}
    </View>
  );
};
