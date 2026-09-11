import {
  AiIcon,
  CheckListIcon,
  ObjectiveIcon,
  RequestsIcon,
  RoadmapIcon,
} from "icons";
import type { MarketingVisual } from "@/components/shared/marketing-detail-page";
import {
  ContextListPreview,
  EvidencePreview,
} from "@/modules/home/capability-previews";
import {
  IntegrationBrand,
  type IntegrationBrandName,
} from "@/modules/integrations/integration-brand";

export type PreviewKind =
  | "work"
  | "feedback"
  | "goals"
  | "roadmap"
  | "planning";
const PREVIEW_ICONS = {
  work: CheckListIcon,
  feedback: RequestsIcon,
  goals: ObjectiveIcon,
  roadmap: RoadmapIcon,
  planning: AiIcon,
};

export function MarketingVisualCard({
  visual,
  kind = "work",
  brand,
}: {
  visual: MarketingVisual;
  kind?: PreviewKind;
  brand?: IntegrationBrandName;
}) {
  const Icon = PREVIEW_ICONS[kind];

  const icon = brand ? (
    <IntegrationBrand name={brand} size={20} surface="light" />
  ) : (
    <Icon className="size-4 text-current" />
  );
  const [request, work] = visual.rows;

  if (!request || !work || visual.rows.length > 2) {
    return (
      <div aria-hidden="true">
        <ContextListPreview
          footer={visual.note ?? visual.subheading}
          items={visual.rows.map((row) => ({
            id: row.label,
            label: row.value,
            detail: row.label,
            icon,
          }))}
          title={visual.heading}
        />
      </div>
    );
  }

  return (
    <div aria-hidden="true">
      <EvidencePreview
        context={request.label}
        footer={visual.note}
        icon={icon}
        note={visual.subheading}
        request={request.value}
        status={visual.badge}
        title={visual.heading}
        workLabel={work.label}
        workTitle={work.value}
      />
    </div>
  );
}
