import type { ReactNode } from "react";
import { useSyncExternalStore } from "react";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { isStoryUuid } from "@/shared/routing/story";
import { parseNativeLink } from "@/components/ui/link-preview/native-target";
import {
  openNativeLink,
  useNativeLinkNavigation,
} from "@/components/ui/link-preview/native-navigation";

const subscribeToOrigin = () => () => undefined;
const getClientOrigin = () => window.location.origin;
const getServerOrigin = () => undefined;

export const MessageLink = ({
  href,
  children,
  shorten = true,
}: {
  href: string;
  children: ReactNode;
  shorten?: boolean;
}) => {
  const { workspaceSlug } = useWorkspacePath();
  const navigate = useNativeLinkNavigation();
  const origin = useSyncExternalStore(
    subscribeToOrigin,
    getClientOrigin,
    getServerOrigin,
  );
  const target = parseNativeLink(href, {
    workspaceSlug,
    origin,
  });
  let label: ReactNode = children;
  if (target && shorten) {
    switch (target.kind) {
      case "story":
        label = isStoryUuid(target.id) ? "Story" : target.id;
        break;
      case "objective":
        label = "Objective";
        break;
      case "document":
        label = "Document";
        break;
      case "unsupported":
        label = "Workspace link";
    }
  }
  return (
    <a
      href={href}
      onClick={(event) => {
        openNativeLink(event, target, navigate);
      }}
      rel="noopener noreferrer"
      target="_blank"
    >
      {label}
    </a>
  );
};
