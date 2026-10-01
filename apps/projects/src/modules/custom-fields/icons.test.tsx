import { render } from "@testing-library/react";
import { CustomFieldIcon, getCustomFieldIconKey } from "./icons";

describe("custom field icon rendering", () => {
  it("uses a real money glyph for automatic money fields without a stored override", () => {
    const field = { type: "money", icon: null } as const;
    expect(getCustomFieldIconKey(field)).toBe("money");
    const { container } = render(<CustomFieldIcon field={field} />);
    const svg = container.querySelector("svg");
    expect(svg).toHaveAttribute("aria-hidden", "true");
    expect(svg).toHaveAttribute("focusable", "false");
    expect(svg).toHaveAttribute("stroke", "currentColor");
    expect(svg?.querySelectorAll("path")).toHaveLength(6);
  });

  it("keeps the saved lock key and supplies its shackle with a visible inherited stroke", () => {
    const field = { type: "text", icon: "lock" } as const;
    expect(getCustomFieldIconKey(field)).toBe("lock");
    const { container } = render(<CustomFieldIcon field={field} />);
    const svg = container.querySelector("svg");
    const shackle = svg?.querySelector("path");
    expect(shackle).not.toBeNull();
    // The shared lock's first path previously had no stroke of its own or
    // inherited from the SVG, making the padlock look like a detached box.
    expect(svg).toHaveAttribute("stroke", "currentColor");
  });
});
