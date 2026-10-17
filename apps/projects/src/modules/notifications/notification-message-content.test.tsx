import { render } from "@testing-library/react";
import type { AppNotification } from "./types";
import { NotificationMessageContent } from "./notification-message-content";
import { renderTemplate } from "./utils/render-template";

describe("NotificationMessageContent", () => {
  it("localizes only whole words in the template without changing notification data", () => {
    const rendered = renderTemplate({
      template: "{actor} updated the story history: {content}. View the story.",
      variables: {
        actor: { value: "storyteller", type: "actor" },
        content: {
          value: "The story of our history page",
          type: "plain_text",
        },
      },
    });
    const view = render(
      <div data-testid="message">
        <NotificationMessageContent
          segments={rendered.segments}
          storyTerm="work item"
        />
      </div>,
    );

    expect(view.getByTestId("message")).toHaveTextContent(
      "storyteller updated the work item history: The story of our history page. View the work item.",
    );
  });

  it("renders untrusted template values as inert text", () => {
    const actor = '<img src=x onerror="actor()">';
    const content =
      '<script>alert("xss")</script><img src=x onerror="alert(1)">&lt;entity&gt;';
    const message: AppNotification["message"] = {
      template: "{actor} commented on the story: {content}",
      variables: {
        actor: { value: actor, type: "actor" },
        content: { value: content, type: "text" },
      },
    };

    const rendered = renderTemplate(message);
    const view = render(
      <div data-testid="message">
        <NotificationMessageContent
          segments={rendered.segments}
          storyTerm="work item"
        />
      </div>,
    );

    expect(view.getByTestId("message")).toHaveTextContent(
      `${actor} commented on the work item: <entity>`,
    );
    expect(view.container.querySelector("script")).not.toBeInTheDocument();
    expect(view.container.querySelector("img")).not.toBeInTheDocument();
    expect(view.container.querySelector("[onerror]")).not.toBeInTheDocument();

    const emphasizedValues = view.container.querySelectorAll(
      ".font-semibold.antialiased",
    );
    expect(emphasizedValues).toHaveLength(1);
    expect(emphasizedValues[0]).toHaveTextContent(actor);
    expect(emphasizedValues[0]).not.toHaveTextContent(content);
  });
});
