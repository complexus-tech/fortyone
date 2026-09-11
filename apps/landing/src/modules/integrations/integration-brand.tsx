import Image from "next/image";
import { cn } from "lib";

const INTEGRATION_BRANDS = {
  Slack: { src: "/integrations/slack.svg", invert: false },
  GitHub: { src: "/integrations/github-mark.svg", invert: true },
  "Google Calendar": {
    src: "/integrations/google-calendar-2026.svg",
    invert: false,
  },
  "Outlook Calendar": { src: "/integrations/outlook-2025.svg", invert: false },
  "Google Drive": { src: "/integrations/drive.svg", invert: false },
  Figma: { src: "/integrations/figma.svg", invert: false },
  ChatGPT: { src: "/integrations/chatgpt-mark.svg", invert: true },
  Claude: { src: "/integrations/claude-color.svg", invert: false },
  Cursor: { src: "/integrations/cursor.svg", invert: true },
} as const;

export type IntegrationBrandName = keyof typeof INTEGRATION_BRANDS;

export function IntegrationBrand({
  name,
  size = 28,
  surface = "theme",
}: {
  name: IntegrationBrandName;
  size?: number;
  surface?: "theme" | "light";
}) {
  const brand = INTEGRATION_BRANDS[name];
  return (
    <Image
      alt={name}
      className={cn(
        "shrink-0 object-contain",
        brand.invert && surface === "theme" && "dark:invert",
      )}
      height={size}
      sizes={`${size}px`}
      src={brand.src}
      width={size}
    />
  );
}
