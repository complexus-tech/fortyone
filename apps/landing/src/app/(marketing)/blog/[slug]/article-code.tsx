"use client";

import type { ComponentPropsWithoutRef } from "react";
import { useRef, useState } from "react";

export function ArticleCode({
  children,
  ...props
}: ComponentPropsWithoutRef<"pre">) {
  const ref = useRef<HTMLPreElement>(null);
  const [message, setMessage] = useState("Copy");
  async function copy() {
    try {
      await navigator.clipboard.writeText(ref.current?.textContent ?? "");
      setMessage("Copied");
    } catch {
      setMessage("Select text to copy");
    }
  }
  return (
    <div data-code-block>
      <div data-code-toolbar>
        <span>Code</span>
        <button
          aria-label="Copy code"
          aria-live="polite"
          onClick={copy}
          type="button"
        >
          {message}
        </button>
      </div>
      <pre ref={ref} {...props}>
        {children}
      </pre>
    </div>
  );
}
