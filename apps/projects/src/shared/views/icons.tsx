import { ViewsIcon } from "icons";
import type { ViewIconKey } from "./metadata";
import { VIEW_ICON_OPTIONS } from "./icon-options";

/** Stored keys select trusted shared SVGs; saved views never supply markup. */
export const SavedViewIcon = ({
  configuration,
  className,
}: {
  configuration: { icon?: ViewIconKey | null };
  className?: string;
}) => {
  const Icon =
    VIEW_ICON_OPTIONS.find(({ value }) => value === configuration.icon)?.Icon ??
    ViewsIcon;
  return <Icon className={className} />;
};
