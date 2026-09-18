import { Button } from "ui";

// The same floating attribution used on the public roadmap and shared docs.
export function PublicBranding({
  href = "https://fortyone.app",
  label = "Powered by FortyOne",
}: {
  href?: string;
  label?: string;
}) {
  return (
    <Button
      className="bg-surface-elevated/90 shadow-shadow fixed right-4 bottom-4 z-30 h-10 border-[0.5px] px-3 shadow-lg backdrop-blur md:right-6 md:bottom-6"
      color="tertiary"
      href={href}
      size="sm"
      variant="outline"
    >
      {label}
    </Button>
  );
}
