import { ChevronLeftIcon } from "icons";
import { Button } from "ui";

type SettingsBackButtonProps = {
  href: string;
  label: string;
};

export const SettingsBackButton = ({
  href,
  label,
}: SettingsBackButtonProps) => (
  <Button
    asIcon
    className="bg-state-hover shrink-0"
    color="tertiary"
    href={href}
    size="sm"
    variant="naked"
  >
    <ChevronLeftIcon aria-hidden="true" className="size-4" />
    <span className="sr-only">{label}</span>
  </Button>
);
