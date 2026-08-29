import { ResourceNotFoundState } from "@/components/ui/resource-not-found-state";

export default function NotFound() {
  return (
    <ResourceNotFoundState
      description="Oops! It seems the objective path hit a snag. Our team’s on it! While we clear the roadblock, why not explore other routes to productivity?"
      href="/"
      title="404: Objective Detour"
    />
  );
}
