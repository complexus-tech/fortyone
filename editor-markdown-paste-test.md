# FortyOne Markdown paste test

Use this document to verify that pasting Markdown into the document editor creates rich content rather than leaving Markdown syntax visible.

## Text formatting

This paragraph includes **bold text**, _italic text_, ~~strikethrough~~, `inline code`, and a [link to FortyOne](https://fortyone.app).

> A good editor keeps structure intact while making the writing experience feel calm and predictable.

---

## Lists and tasks

- Keep project context close to execution.
- Make decisions and risks easy to scan.
  - Preserve nested list structure.
  - Keep indentation clear.
- Support clear ownership and next steps.

1. Paste this entire file into a blank document.
2. Confirm headings and text styles render correctly.
3. Confirm tables, code, and task lists remain structured.

- [x] Headings render as headings
- [x] Inline formatting is preserved
- [ ] Table cells are editable
- [ ] Public preview matches the editor

## Workstream status

| Workstream        | Owner       | Status      | Next milestone               |
| :---------------- | :---------- | :---------- | :--------------------------- |
| Research          | Product     | Complete    | Share findings               |
| Editor experience | Design      | In progress | Review interaction model     |
| Delivery          | Engineering | Planned     | Confirm implementation scope |

## Alignment test

| Item            | Quantity | Progress |
| :-------------- | -------: | :------: |
| Editor polish   |        4 |   75%    |
| Export formats  |        3 |   100%   |
| Comment threads |        2 |   50%    |

## Code

```ts
type DocumentStatus = "draft" | "published";

const getDocumentLabel = (title: string, status: DocumentStatus) =>
  `${title} (${status})`;
```

## Final checks

### Images and media

Images pasted separately should use the same corner shape and spacing in both the editor and public preview.

### Expected result

No Markdown markers should remain visible after pasting, tables should be real editable tables, and the document should keep the same visual hierarchy in the editor and public preview.
