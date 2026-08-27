/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolSet } from "ai";
import { getRequestOptionsScope } from "@/lib/http/request-options-scope";
import {
  runWithMayaHttpRequestContext,
  withMayaHttpRequestContext,
} from "./maya-http-request-context";

jest.mock("server-only", () => ({}));

describe("Maya HTTP request context", () => {
  it("retains the request signal across asynchronous work", async () => {
    const signal = new AbortController().signal;

    const observedSignal = await runWithMayaHttpRequestContext(
      signal,
      async () => {
        await Promise.resolve();
        return getRequestOptionsScope()?.signal;
      },
    );

    expect(observedSignal).toBe(signal);
    expect(getRequestOptionsScope()).toBeUndefined();
  });

  it("isolates concurrent Maya requests", async () => {
    const firstSignal = new AbortController().signal;
    const secondSignal = new AbortController().signal;

    const observedSignals = await Promise.all([
      runWithMayaHttpRequestContext(firstSignal, async () => {
        await Promise.resolve();
        return getRequestOptionsScope()?.signal;
      }),
      runWithMayaHttpRequestContext(secondSignal, async () => {
        await Promise.resolve();
        return getRequestOptionsScope()?.signal;
      }),
    ]);

    expect(observedSignals).toEqual([firstSignal, secondSignal]);
  });

  it("restores the outer request context after nested execution", () => {
    const outerSignal = new AbortController().signal;
    const innerSignal = new AbortController().signal;

    runWithMayaHttpRequestContext(outerSignal, () => {
      expect(getRequestOptionsScope()?.signal).toBe(outerSignal);
      runWithMayaHttpRequestContext(innerSignal, () => {
        expect(getRequestOptionsScope()?.signal).toBe(innerSignal);
      });
      expect(getRequestOptionsScope()?.signal).toBe(outerSignal);
    });
  });

  it("scopes each tool execution to the AI SDK abort signal", async () => {
    const signal = new AbortController().signal;
    const execute = jest.fn(
      async (_input: unknown, _options: ToolExecutionOptions) => {
        await Promise.resolve();
        return getRequestOptionsScope()?.signal;
      },
    );
    const wrappedTools = withMayaHttpRequestContext({
      testTool: { execute },
    } as unknown as ToolSet);

    const observedSignal = await wrappedTools.testTool.execute?.(
      {},
      {
        abortSignal: signal,
        messages: [],
        toolCallId: "tool-call-1",
      },
    );

    expect(observedSignal).toBe(signal);
    expect(execute).toHaveBeenCalledTimes(1);
    expect(getRequestOptionsScope()).toBeUndefined();
  });
});
