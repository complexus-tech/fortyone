import { CloseIcon } from "icons";
import { cn } from "lib";
import type { GoogleDriveFileContext } from "@/lib/ai/google-drive-context";
import { GoogleFileTypeIcon } from "@/modules/google-drive/google-file-type-icon";

export const GoogleDriveContextChips = ({
  files,
  onRemove,
  className,
}: {
  files: GoogleDriveFileContext[];
  onRemove?: (referenceId: string) => void;
  className?: string;
}) => {
  if (files.length === 0) return null;

  return (
    <div className={cn("flex flex-wrap gap-2", className)}>
      {files.map((file) => (
        <span
          className="border-border/70 bg-surface-muted/10 text-text-primary flex max-w-full items-center gap-2 rounded-xl border px-3 py-2.5 text-sm"
          key={file.referenceId}
          title={file.name}
        >
          <GoogleFileTypeIcon
            className="size-7 bg-transparent"
            mimeType={file.mimeType}
          />
          <span className="max-w-56 truncate">{file.name}</span>
          {onRemove ? (
            <button
              aria-label={`Remove ${file.name}`}
              className="text-text-muted hover:text-text-primary focus-visible:ring-ring/50 -mr-1 flex size-6 shrink-0 items-center justify-center rounded-md transition-colors outline-none focus-visible:ring-2"
              onClick={() => {
                onRemove(file.referenceId);
              }}
              type="button"
            >
              <CloseIcon aria-hidden="true" className="size-3.5" />
            </button>
          ) : null}
        </span>
      ))}
    </div>
  );
};
