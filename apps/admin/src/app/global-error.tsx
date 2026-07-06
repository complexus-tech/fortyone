"use client";

import "ui/styles.css";
import "../styles/global.css";
import { AdminErrorState } from "@/components/admin-error-state";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html lang="en">
      <body>
        <AdminErrorState digest={error.digest} onRetry={reset} />
      </body>
    </html>
  );
}
