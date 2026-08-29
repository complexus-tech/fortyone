"use client";

import { AdminErrorState } from "@/components/admin-error-state";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return <AdminErrorState digest={error.digest} onRetry={reset} />;
}
