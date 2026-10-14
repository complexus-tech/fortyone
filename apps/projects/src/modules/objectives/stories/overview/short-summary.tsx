import { useLayoutEffect, useRef } from "react";

type ShortSummaryProps = {
  value: string;
  readOnly: boolean;
  onChange: (value: string) => void;
  onBlur: () => void;
};

export const ShortSummary = ({
  value,
  readOnly,
  onChange,
  onBlur,
}: ShortSummaryProps) => {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useLayoutEffect(() => {
    const textarea = textareaRef.current;
    if (!textarea) return;

    const fitContent = () => {
      textarea.style.height = "auto";
      textarea.style.height = `${textarea.scrollHeight}px`;
    };
    fitContent();

    // Reflow when the resizable objective panel changes width, without responding
    // to the height adjustments made here.
    let width = textarea.clientWidth;
    const observer = new ResizeObserver(() => {
      if (textarea.clientWidth === width) return;
      width = textarea.clientWidth;
      fitContent();
    });
    observer.observe(textarea);
    return () => {
      observer.disconnect();
    };
  }, [value]);

  return (
    <textarea
      aria-label="Objective short summary"
      className="text-foreground placeholder:text-text-muted mt-3 block w-full resize-none overflow-hidden border-0 bg-transparent p-0 leading-6 focus-visible:outline-0"
      maxLength={500}
      onBlur={onBlur}
      onChange={(event) => {
        onChange(event.target.value);
      }}
      placeholder="Add short summary..."
      readOnly={readOnly}
      ref={textareaRef}
      rows={1}
      value={value}
    />
  );
};
