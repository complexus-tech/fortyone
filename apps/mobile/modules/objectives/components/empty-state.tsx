import { QueryState } from "@/components/ui/query-state";
import { useTerminology } from "@/hooks/use-terminology";

export const EmptyState = ({
  title,
  message,
}: {
  title?: string;
  message?: string;
}) => {
  const { getTermDisplay } = useTerminology();
  return (
    <QueryState
      title={
        title ||
        `No ${getTermDisplay("objectiveTerm", { variant: "plural" })} yet`
      }
      message={
        message || "When your team has work here, it will appear in this list."
      }
    />
  );
};
