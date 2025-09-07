import type { BaseLayoutProps } from "fumadocs-ui/layouts/shared";
import { cn } from "lib";

/**
 * Shared layout configurations
 *
 * you can customise layouts individually from:
 * Home Layout: app/(home)/layout.tsx
 * Docs Layout: app/docs/layout.tsx
 */
export const baseOptions: BaseLayoutProps = {
  nav: {
    title: (
      <div
        className={cn(
          "text-[1.5rem] font-semibold text-black dark:text-white font-(family-name:--font-heading)"
        )}
      >
        forty
        <span className="ml-0.5 inline-block bg-[#000000] px-0.5 pb-0.5 leading-none text-white dark:bg-white dark:text-black">
          one
        </span>
      </div>
    ),
  },
  // see https://fumadocs.dev/docs/ui/navigation/links
  links: [],
};
