import { MainLogo } from "@/components/logo";
import type { BaseLayoutProps } from "fumadocs-ui/layouts/shared";
import { BlogIcon, DocsIcon } from "icons";

/**
 * Shared layout configurations
 *
 * you can customise layouts individually from:
 * Home Layout: app/(home)/layout.tsx
 * Docs Layout: app/docs/layout.tsx
 */
export const baseOptions: BaseLayoutProps = {
  nav: {
    title: <MainLogo className="h-6 relative -left-3.5" />,
  },
  // see https://fumadocs.dev/docs/ui/navigation/links
  links: [],
};
