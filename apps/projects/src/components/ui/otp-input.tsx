"use client";

import {
  forwardRef,
  useEffect,
  useRef,
  useState,
  type ClipboardEvent,
  type KeyboardEvent,
} from "react";
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
    const inputSlots = Array.from({ length }, (_, index) => ({
      id: `otp-slot-${index + 1}`,
      index,
    }));

    // Initialize input refs array
    useEffect(() => {
      inputRefs.current = inputRefs.current.slice(0, length);
    }, [length]);

    const handleInputChange = (index: number, inputValue: string) => {
      // Only allow single digit
      const digit = inputValue.slice(-1);

      // Only allow digits
      if (!/^\d*$/.test(digit)) {
        return;
      }

      const newValue = value.split("");
      newValue[index] = digit;
      const updatedValue = newValue.join("").padEnd(length, "");

      onChange(updatedValue);

      // Auto-focus next input
      if (digit && index < length - 1) {
        inputRefs.current[index + 1]?.focus();
      }
    };

    const handleKeyDown = (index: number, e: KeyboardEvent) => {
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

    const handlePaste = (e: ClipboardEvent) => {
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
      <div className={cn("flex gap-2", className)} ref={ref}>
        {inputSlots.map((slot) => (
          <input
            aria-label={`Verification code digit ${slot.index + 1} of ${length}`}
            className={cn(
              "size-12 rounded-lg border text-center text-lg font-semibold transition-all duration-200",
              "border-border bg-surface/70",
              "focus:ring-border dark:focus:ring-border focus:ring-[2.5px] focus:ring-offset-1 focus:outline-0",
              "disabled:cursor-not-allowed disabled:opacity-50",
              {
                "border-danger focus:ring-danger dark:border-danger dark:focus:ring-danger":
                  hasError,
                "ring-border dark:ring-border ring-[2.5px]":
                  focusedIndex === slot.index,
              },
            )}
            disabled={disabled}
            inputMode="numeric"
            key={slot.id}
            maxLength={1}
            onBlur={() => {
              setFocusedIndex(null);
            }}
            onChange={(event) => {
              handleInputChange(slot.index, event.target.value);
            }}
            onFocus={() => {
              setFocusedIndex(slot.index);
            }}
            onKeyDown={(event) => {
              handleKeyDown(slot.index, event);
            }}
            onPaste={handlePaste}
            pattern="[0-9]*"
            ref={(element) => {
              inputRefs.current[slot.index] = element;
            }}
            type="text"
            value={value[slot.index] || ""}
          />
        ))}
      </div>
    );
  },
);

OTPInput.displayName = "OTPInput";
