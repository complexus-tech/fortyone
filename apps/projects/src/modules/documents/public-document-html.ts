import sanitizeHtml, { defaults, simpleTransform } from "sanitize-html";

// Public document content never renders workspace chrome, relationships, member
// details, history, or unsanitized HTML. Media stays behind the revocable token.
export function publicDocumentHTML(
  html: string,
  token: string,
  apiUrl: string,
) {
  return sanitizeHtml(html, {
    allowedTags: [...defaults.allowedTags, "img", "video", "input"],
    allowedAttributes: {
      a: ["href", "title", "rel", "target"],
      img: [
        "src",
        "alt",
        "title",
        "width",
        "data-width",
        "data-align",
        "style",
      ],
      video: ["src", "controls", "preload"],
      input: ["type", "checked", "disabled"],
      td: ["colspan", "rowspan"],
      th: ["colspan", "rowspan"],
      ul: ["data-type"],
      li: ["data-type", "data-checked"],
    },
    allowedStyles: {
      img: {
        width: [/^\d+(?:\.\d+)?(?:px|%)$/],
        "max-width": [/^100%$/],
        display: [/^block$/],
        "margin-left": [/^(?:auto|0)$/],
        "margin-right": [/^(?:auto|0)$/],
      },
    },
    allowedSchemes: ["https", "http", "mailto"],
    allowProtocolRelative: false,
    transformTags: {
      a: simpleTransform("a", {
        rel: "noopener noreferrer",
        target: "_blank",
      }),
      input: (_tag, attrs) => ({
        tagName: "input",
        attribs: {
          type: "checkbox",
          disabled: "",
          ...(Object.hasOwn(attrs, "checked") ? { checked: "" } : {}),
        },
      }),
      img: (tagName, attribs) => ({
        tagName,
        attribs: rewriteMedia(attribs, token, apiUrl),
      }),
      video: (tagName, attribs) => ({
        tagName,
        attribs: {
          ...rewriteMedia(attribs, token, apiUrl),
          controls: "",
          preload: "metadata",
        },
      }),
    },
  });
}

function rewriteMedia(
  attributes: Record<string, string>,
  token: string,
  apiUrl: string,
) {
  const source = attributes.src;
  const match =
    /\/workspaces\/[^/]+\/documents\/[0-9a-f-]+\/media\/[0-9a-f-]{36}(?=[?#]|$)/.exec(
      source,
    );
  return match
    ? {
        ...attributes,
        src: `${apiUrl}/public/documents/${token}/media/${match[0].slice(-36)}`,
      }
    : attributes;
}
