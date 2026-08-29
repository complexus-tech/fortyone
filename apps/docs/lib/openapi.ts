import { loadOpenAPIContract } from "@/lib/openapi-contract";
import { createOpenAPI } from "fumadocs-openapi/server";

export const openapi = createOpenAPI({
  input: {
    fortyone: loadOpenAPIContract,
  },
});
