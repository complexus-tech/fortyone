import { LoadingIcon } from "icons";
import { Button } from "ui";

export const PreviewLoading = ({
  label = "Loading preview…",
}: {
  label?: string;
}) => (
  <div
    className="text-text-muted flex min-h-48 flex-1 items-center justify-center gap-2 p-6"
    role="status"
  >
    <LoadingIcon
      aria-hidden
      className="size-4 animate-spin motion-reduce:animate-none"
    />
    <span>{label}</span>
  </div>
);

export const PreviewError = ({
  message,
  onRetry,
}: {
  message: string;
  onRetry?: () => void;
}) => (
  <div
    className="flex min-h-48 flex-1 flex-col items-center justify-center gap-3 p-6 text-center"
    role="alert"
  >
    <p className="text-text-muted max-w-md">{message}</p>
    {onRetry ? (
      <Button
        color="tertiary"
        onClick={onRetry}
        size="sm"
        type="button"
        variant="outline"
      >
        Retry preview
      </Button>
    ) : null}
  </div>
);
