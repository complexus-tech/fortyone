import "server-only";

import { AsyncLocalStorage } from "node:async_hooks";
import type { ToolExecutionOptions, ToolSet } from "ai";
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
): TOOLS =>
  Object.fromEntries(
    Object.entries(toolSet).map(([name, registeredTool]) => {
      const execute = registeredTool.execute as
        | NonNullable<ToolSet[string]["execute"]>
        | undefined;
      if (!execute) return [name, registeredTool];

      return [
        name,
        {
          ...registeredTool,
          execute: (input: unknown, options: ToolExecutionOptions) => {
            if (!options.abortSignal) return execute(input, options);

            return runWithMayaHttpRequestContext(options.abortSignal, () =>
              execute(input, options),
            );
          },
        },
      ];
    }),
  ) as unknown as TOOLS;
