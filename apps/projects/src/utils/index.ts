export { getApiError } from "./api-error";
export { hexToRgba } from "./color";
export {
  getRedirectUrl,
  buildWorkspaceUrl,
  withWorkspacePath,
} from "./workspace-url";

export const slugify = (text = "") => {
  return text
    .toString()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase()
    .trim()
    .replace(/\s+/g, "-")
    .replace(/&/g, "-and-")
    .replace(/[^\w-]+/g, "")
    .replace(/--+/g, "-");
};

export const toTitleCase = (text: string) => {
  return text.charAt(0).toUpperCase() + text.slice(1);
};
