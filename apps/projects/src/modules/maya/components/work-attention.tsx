import Link from "next/link";
import { Box, Button, Skeleton } from "ui";
import { cn } from "lib";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useWorkAttention } from "../hooks/use-work-attention";

export const WorkAttention = () => {
  const { withWorkspace } = useWorkspacePath();
  const { getTermDisplay } = useTerminology();
  const { items, isPending, isError, retry } = useWorkAttention();
  if (isPending)
    return (
      <Skeleton aria-label="Loading due dates" className="h-5 w-48 rounded" />
    );
  if (isError)
    return (
      <Box
        className="text-text-muted flex items-center gap-2 text-sm"
        role="status"
      >
        Couldn&apos;t load due dates.
        <Button color="tertiary" onClick={retry} size="sm" variant="naked">
          Retry
        </Button>
      </Box>
    );
  return (
    <Box
      aria-label="Work needing attention"
      className="text-text-muted flex flex-wrap items-center gap-2 text-base"
    >
      {items.map(({ view, count }, index) => (
        <Box className="flex items-center gap-2" key={view}>
          {index > 0 ? <span aria-hidden="true">·</span> : null}
          <Link
            aria-label={`${count} ${getTermDisplay("storyTerm", { variant: "plural" })} ${view === "today" ? "due today" : "overdue"}`}
            className={cn(
              "underline-offset-4 hover:underline focus-visible:underline",
              view === "overdue" && Boolean(count) && "text-danger",
            )}
            href={withWorkspace(`/my-work?tab=assigned&attention=${view}`)}
          >
            {count} {view === "today" ? "due today" : "overdue"}
          </Link>
        </Box>
      ))}
    </Box>
  );
};
