import { Linking } from "react-native";
import { toast } from "sonner-native";

export const publicPages = {
  privacy: { label: "Privacy policy", url: "https://fortyone.app/privacy" },
  terms: { label: "Terms of service", url: "https://fortyone.app/terms" },
  support: { label: "Support", url: "https://fortyone.app/contact" },
} as const;

export const openPublicPage = (page: keyof typeof publicPages) => {
  const { label, url } = publicPages[page];
  void Linking.openURL(url).catch(() => {
    toast.error(`Could not open ${label.toLowerCase()}`, {
      description: "Please try again when you are connected.",
    });
  });
};
