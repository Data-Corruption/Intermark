# Formatting Example

## Introduction

This document is designed to display various Markdown features. Below are examples of different elements you might use. <3

## Code Blocks

### Inline Code

Use the `printf()` function.

### Block Code

```python
def hello_world():
  print("Hello, world!")
```

```javascript
function helloWorld() {
  console.log("Hello, world!");
}
```

## Lists

### Ordered List

1. First item
2. Second item
3. Third item

### Unordered List

- Item 1
- Item 2
  - Subitem 1
  - Subitem 2
- Item 3

## Text Formatting

- _Italic_
- **Bold**
- **_Bold and Italic_**
- ~~Strikethrough~~
- Superscript: X^2^
- Subscript: H~2~O

## Links and Images

### Links

[Google](https://www.google.com)

### Images

![Google Logo](https://www.google.com/favicon.ico)

## Divider

---

## Custom HTML

<div style="text-align: center; font-size: 18px; color: #555;">This text is centered using HTML.</div>

## Tailwindcss / DaisyUI Support

<div class="collapse bg-base-200">
  <input type="checkbox" />
  <div class="collapse-title text-xl font-medium">Click me to show/hide content</div>
  <div class="collapse-content">
    <mdsrc>
# Nested Markdown Example
Anything inside these "mdsrc" tags will be converted to html along with root level markdown.
    </mdsrc>
  </div>
</div>

## Tables

| Syntax    | Description |
| --------- | ----------- |
| Header    | Title       |
| Paragraph | Text        |

## Blockquotes

> This is a blockquote.
>
> It can span multiple lines.
