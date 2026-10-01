import type { ReactNode } from "react";
import type { TeamSettingsSection as Section } from "../navigation";
import { getTeamSettingsSectionId } from "../navigation";

export const TeamSettingsSection = ({
  value,
  label,
  children,
}: {
  value: Section;
  label: string;
  children: ReactNode;
}) => (
  <section
    aria-label={label}
    className="scroll-mt-6 outline-none"
    id={getTeamSettingsSectionId(value)}
    tabIndex={-1}
  >
    {children}
  </section>
);
