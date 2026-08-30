/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { AppNotification } from "../types";
import { renderTemplate } from "./render-template";

describe("renderTemplate", () => {
  it("renders known values while preserving unknown placeholders", () => {
    const message: AppNotification["message"] = {
      template: "{actor} mentioned you in {story}.",
      variables: {
        actor: { value: "Ava" },
      },
    };

    expect(renderTemplate(message)).toEqual({
      text: "Ava mentioned you in {story}.",
      html: '<span class="font-semibold antialiased text-foreground/90">Ava</span> mentioned you in {story}.',
    });
  });
});
