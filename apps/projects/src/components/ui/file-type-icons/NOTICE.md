# File type icon assets

## Microsoft Fluent UI System Icons

Copyright (c) 2020 Microsoft Corporation. Licensed under the MIT License;
see [LICENSE.fluent.txt](./LICENSE.fluent.txt) for the complete license.

Source: https://github.com/microsoft/fluentui-system-icons
Pinned revision: `a347af95478f989038cccbf085978865264c048a`.

`fluent-paths.ts` vendors the unchanged path geometry from these upstream files.
The original monochrome fill is replaced by `currentColor` in `FileTypeIcon`,
so the glyphs use the application's tile colors.

- `csv`: https://github.com/microsoft/fluentui-system-icons/blob/a347af95478f989038cccbf085978865264c048a/assets/Table/SVG/ic_fluent_table_24_regular.svg
- `text`: https://github.com/microsoft/fluentui-system-icons/blob/a347af95478f989038cccbf085978865264c048a/assets/Document%20Text/SVG/ic_fluent_document_text_24_regular.svg
- `image`: https://github.com/microsoft/fluentui-system-icons/blob/a347af95478f989038cccbf085978865264c048a/assets/Image/SVG/ic_fluent_image_24_regular.svg
- `video`: https://github.com/microsoft/fluentui-system-icons/blob/a347af95478f989038cccbf085978865264c048a/assets/Video%20Clip/SVG/ic_fluent_video_clip_24_regular.svg
- `audio`: https://github.com/microsoft/fluentui-system-icons/blob/a347af95478f989038cccbf085978865264c048a/assets/Speaker%202/SVG/ic_fluent_speaker_2_24_regular.svg
- `archive`: https://github.com/microsoft/fluentui-system-icons/blob/a347af95478f989038cccbf085978865264c048a/assets/Archive/SVG/ic_fluent_archive_24_regular.svg
- `code`: https://github.com/microsoft/fluentui-system-icons/blob/a347af95478f989038cccbf085978865264c048a/assets/Code/SVG/ic_fluent_code_24_regular.svg
- `link`: https://github.com/microsoft/fluentui-system-icons/blob/a347af95478f989038cccbf085978865264c048a/assets/Globe/SVG/ic_fluent_globe_24_regular.svg
- `file`: https://github.com/microsoft/fluentui-system-icons/blob/a347af95478f989038cccbf085978865264c048a/assets/Document/SVG/ic_fluent_document_24_regular.svg

## Office file icons from vscode-icons

Source: https://github.com/vscode-icons/vscode-icons
Pinned revision: `82dde38f58ffc2a7def3efadb11208d010841a8f`.

`office-glyphs.tsx` incorporates these established Word, Excel, and PowerPoint
SVGs, retaining their original geometry and colors. JSX attribute names and
per-instance gradient IDs are the only adaptations.

- https://github.com/vscode-icons/vscode-icons/blob/82dde38f58ffc2a7def3efadb11208d010841a8f/icons/file_type_word.svg
- https://github.com/vscode-icons/vscode-icons/blob/82dde38f58ffc2a7def3efadb11208d010841a8f/icons/file_type_excel.svg
- https://github.com/vscode-icons/vscode-icons/blob/82dde38f58ffc2a7def3efadb11208d010841a8f/icons/file_type_powerpoint.svg

The upstream project distinguishes its source-code license (MIT), icon license
[Creative Commons Attribution-ShareAlike 4.0](https://creativecommons.org/licenses/by-sa/4.0/),
and branded icons, which remain subject to their respective copyright licenses:
https://github.com/vscode-icons/vscode-icons/blob/82dde38f58ffc2a7def3efadb11208d010841a8f/README.md#license.

Microsoft Word, Excel, and PowerPoint names and branded artwork belong to
Microsoft Corporation. They are used here solely to identify the corresponding
file types; this application does not claim ownership or Microsoft endorsement.
The branded assets are not relicensed under the application's license.

## PDF

The existing application PDF glyph is retained in `../file-type-icon.tsx`.
It is not copied from either of the above collections.
