"use client";
import { FormEvent, HTMLProps, useEffect, useRef } from "react";

import { cn } from "lib";
import { Box } from "../Box/Box";
import { Text } from "../Text/Text";

interface ContentEditableProps extends HTMLProps<HTMLDivElement> {
  placeholder?: string;
  value: string;
  setValue: (value: string) => void;
}

export const ContentEditable = ({
  placeholder,
  className,
  value = "",
  setValue,
}: ContentEditableProps) => {
  const contentRef = useRef<HTMLDivElement>(null);

  const handleChange = (e: FormEvent<HTMLDivElement>) => {
    setValue(e.currentTarget.textContent || "");
  };

  const sanitizeHTML = (input: string): string => {
    const doc = new DOMParser().parseFromString(input, "text/html");
    doc.querySelectorAll("script").forEach((script) => script.remove());
    doc.querySelectorAll("style").forEach((style) => style.remove());
    return new XMLSerializer().serializeToString(doc.body);
  };

  useEffect(() => {
    if (contentRef.current?.textContent !== value) {
      if (contentRef.current) {
        contentRef.current.textContent = sanitizeHTML(value);
      }
    }
  }, [value]);

  return (
    <Box className="relative">
      <div
        className={cn("relative outline-none", className)}
        contentEditable
        onInput={handleChange}
        onKeyDown={handleChange}
        onKeyUp={handleChange}
        onBlur={handleChange}
        ref={contentRef}
      />
      {!contentRef.current?.textContent && (
        <Text
          color="muted"
          className={cn("absolute inset-0 pointer-events-none", className)}
        >
          {placeholder}
        </Text>
      )}
    </Box>
  );
};
