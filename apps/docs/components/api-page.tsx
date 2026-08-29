"use client";

import { createCodeUsageGeneratorRegistry } from "fumadocs-openapi/requests/generators";
import { registerDefault } from "fumadocs-openapi/requests/generators/all";
import { createOpenAPIPage } from "fumadocs-openapi/ui";

const codeUsages = createCodeUsageGeneratorRegistry();
registerDefault(codeUsages);

export const OpenAPIPage = createOpenAPIPage({
  codeUsages,
});
