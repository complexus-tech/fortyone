"use client";

import { forwardRef, useRef, useEffect, useState } from "react";
import { cn } from "lib";

export interface OTPInputProps {
  value: string;
  onChange: (value: string) => void;
  length?: number;
  className?: string;
  hasError?: boolean;
  disabled?: boolean;
}

export const OTPInput = forwardRef<HTMLDivElement, OTPInputProps>(
  ({ value, onChange, length = 6, className, hasError, disabled }, ref) => {
    const inputRefs = useRef<(HTMLInputElement | null)[]>([]);
    const [focusedIndex, setFocusedIndex] = useState<number | null>(null);

    // Initialize input refs array
    useEffect(() => {
      inputRefs.current = inputRefs.current.slice(0, length);
    }, [length]);

    const handleInputChange = (index: number, inputValue: string) => {
      // Only allow single digit
      if (inputValue.length > 1) {
        inputValue = inputValue.slice(-1);
      }

      // Only allow digits
      if (!/^\d*$/.test(inputValue)) {
        return;
      }

      const newValue = value.split("");
      newValue[index] = inputValue;
      const updatedValue = newValue.join("").padEnd(length, "");

      onChange(updatedValue);

      // Auto-focus next input
      if (inputValue && index < length - 1) {
        inputRefs.current[index + 1]?.focus();
      }
    };

    const handleKeyDown = (index: number, e: React.KeyboardEvent) => {
      if (e.key === "Backspace") {
        if (!value[index] && index > 0) {
          // If current input is empty, focus previous input
          inputRefs.current[index - 1]?.focus();
        }
      } else if (e.key === "ArrowLeft" && index > 0) {
        inputRefs.current[index - 1]?.focus();
      } else if (e.key === "ArrowRight" && index < length - 1) {
        inputRefs.current[index + 1]?.focus();
      }
    };

    const handlePaste = (e: React.ClipboardEvent) => {
      e.preventDefault();
      const pastedData = e.clipboardData.getData("text").replace(/\D/g, "");
      if (pastedData.length <= length) {
        onChange(pastedData.padEnd(length, ""));
        // Focus the last filled input or the first empty one
        const lastFilledIndex = Math.min(pastedData.length - 1, length - 1);
        inputRefs.current[lastFilledIndex]?.focus();
      }
    };

    return (
      <div ref={ref} className={cn("flex gap-2", className)}>
        {Array.from({ length }, (_, index) => (
          <input
            key={index}
            ref={(el) => (inputRefs.current[index] = el)}
            type="text"
            inputMode="numeric"
            pattern="[0-9]*"
            maxLength={1}
            value={value[index] || ""}
            onChange={(e) => handleInputChange(index, e.target.value)}
            onKeyDown={(e) => handleKeyDown(index, e)}
            onPaste={handlePaste}
            onFocus={() => setFocusedIndex(index)}
            onBlur={() => setFocusedIndex(null)}
            disabled={disabled}
            className={cn(
              "size-12 rounded-[0.6rem] border text-center text-lg font-semibold transition-all duration-200",
              "border-gray-100 bg-white/70 dark:border-dark-100 dark:bg-dark/20",
              "focus:outline-0 focus:ring-[2.5px] focus:ring-gray-100 focus:ring-offset-1 dark:focus:ring-dark-50",
              "disabled:cursor-not-allowed disabled:opacity-50",
              {
                "border-danger focus:ring-danger dark:border-danger dark:focus:ring-danger":
                  hasError,
                "ring-[2.5px] ring-gray-100 dark:ring-dark-50":
                  focusedIndex === index,
              },
            )}
          />
        ))}
      </div>
    );
  },
);

OTPInput.displayName = "OTPInput";
