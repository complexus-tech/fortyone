import { Badge } from "ui";
import { humanizeKey } from "@/lib/format";

const attentionStates = new Set([
  "degraded",
  "failed",
  "reauthorization_required",
  "suspended",
]);

export const IntegrationStatusBadge = ({ state }: { state: string }) => {
  if (attentionStates.has(state)) {
    return (
      <Badge color="danger" variant="outline">
        {humanizeKey(state)}
      </Badge>
    );
  }

  return <Badge color="tertiary">{humanizeKey(state)}</Badge>;
};
