import "server-only";

import { AsyncLocalStorage } from "node:async_hooks";
import type {
  InferToolInput,
  InferToolOutput,
  Tool,
  ToolExecutionOptions,
  ToolSet,
} from "ai";
import { z } from "zod";
import { googleDriveFileContextSchema } from "@/lib/ai/google-drive-context";
import {
  installRequestOptionsScopeResolver,
  type RequestOptionsScope,
} from "@/lib/http/request-options-scope";

const MAYA_HTTP_REQUEST_CONTEXT_STORAGE_KEY = Symbol.for(
  "fortyone.maya-http-request-context-storage",
);

type GlobalWithMayaHttpRequestContextStorage = typeof globalThis & {
  [MAYA_HTTP_REQUEST_CONTEXT_STORAGE_KEY]?: AsyncLocalStorage<RequestOptionsScope>;
};

const mayaHttpRequestContextStorage = (() => {
  const globalState = globalThis as GlobalWithMayaHttpRequestContextStorage;
  const existing = globalState[MAYA_HTTP_REQUEST_CONTEXT_STORAGE_KEY];
  if (existing) return existing;

  const storage = new AsyncLocalStorage<RequestOptionsScope>();
  globalState[MAYA_HTTP_REQUEST_CONTEXT_STORAGE_KEY] = storage;
  return storage;
})();

installRequestOptionsScopeResolver(() =>
  mayaHttpRequestContextStorage.getStore(),
);

export const mayaToolContextSchema = z.object({
  chatId: z.string(),
  selectedGoogleDriveFiles: z.array(googleDriveFileContextSchema).default([]),
  workspaceSlug: z.string(),
});

export type MayaToolContext = z.infer<typeof mayaToolContextSchema>;

type MayaContextToolSet<TOOLS extends ToolSet> = {
  [NAME in keyof TOOLS]: Tool<
    InferToolInput<TOOLS[NAME]>,
    InferToolOutput<TOOLS[NAME]>,
    MayaToolContext
  >;
};

export const runWithMayaHttpRequestContext = <T>(
  signal: AbortSignal,
  callback: () => T,
): T => mayaHttpRequestContextStorage.run({ signal }, callback);

/**
 * Scope only the disposable HTTP work performed by a model-invoked tool.
 * Transcript finalization and mutation-ledger persistence deliberately remain
 * outside this context so they can finish after a browser disconnect.
 */
export const withMayaHttpRequestContext = <TOOLS extends ToolSet>(
  toolSet: TOOLS,
): MayaContextToolSet<TOOLS> =>
  Object.fromEntries(
    Object.entries(toolSet).map(([name, registeredTool]) => {
      const execute = registeredTool.execute as
        | NonNullable<ToolSet[string]["execute"]>
        | undefined;
      if (!execute)
        return [
          name,
          { ...registeredTool, contextSchema: mayaToolContextSchema },
        ];

      return [
        name,
        {
          ...registeredTool,
          contextSchema: mayaToolContextSchema,
          execute: (
            input: unknown,
            options: ToolExecutionOptions<MayaToolContext>,
          ) => {
            if (!options.abortSignal) return execute(input, options);

            return runWithMayaHttpRequestContext(options.abortSignal, () =>
              execute(input, options),
            );
          },
        },
      ];
    }),
  ) as unknown as MayaContextToolSet<TOOLS>;

export const createMayaToolsContext = <TOOLS extends ToolSet>(
  toolSet: TOOLS,
  context: MayaToolContext,
) =>
  Object.fromEntries(Object.keys(toolSet).map((name) => [name, context])) as {
    [NAME in keyof TOOLS]: MayaToolContext;
  };
