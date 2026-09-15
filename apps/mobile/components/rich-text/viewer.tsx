import { Linking } from "react-native";
import { toast } from "sonner-native";
import { useTheme } from "@/hooks";
import { isSafeLink } from "./content";
import RichTextViewerDOM from "./viewer.dom";

export const RichTextViewer = ({ html }: { html: string }) => {
  const { resolvedTheme } = useTheme();
  if (!html) return null;
  return (
    <RichTextViewerDOM
      html={html}
      dark={resolvedTheme === "dark"}
      onOpenLink={async (href) => {
        if (href.startsWith("/profile/")) {
          toast.info("View this person's profile in the web app.");
          return;
        }
        if (!isSafeLink(href)) return;
        try {
          await Linking.openURL(href);
        } catch {
          toast.error("Could not open this link.");
        }
      }}
      dom={{
        matchContents: true,
        scrollEnabled: false,
        containerStyle: { width: "100%", flex: 0 },
        style: { backgroundColor: "transparent", minHeight: 1 },
      }}
    />
  );
};
