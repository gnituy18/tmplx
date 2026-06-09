# Code snippets (tree-sitter highlighting)

tmplx code is Go embedded in HTML, so no single highlighter handles it. This
renders snippets the same way `tmplx.nvim` does: the **HTML** tree-sitter
grammar for the file, with the **Go** grammar injected into the
`<script type="text/tmplx">` body.

## How it works

`build.mjs` reads two source roots and writes one fragment per file to
`../assets/snippets/<name>.html`:

- `../components/*.html` — the real tmplx components (used by `/examples/`)
- `src/**` — standalone snippets (the docs code blocks live in `src/docs/`)

Each file is highlighted by extension:

- **`.html`** — highlight the whole file as HTML, then (if it has a
  `<script type="text/tmplx">`) highlight just that body as **Go** and splice
  the Go-highlighted lines back in. This is the same HTML+Go injection
  `tmplx.nvim` uses.
- **`.go`** — highlighted as Go.
- **`.sh` / `.txt`** — emitted as plain escaped text (shell commands etc.).

Fragments are `<pre><code>` with CSS classes — no inline styles, no
hand-escaping (the source is the real file). The token theme is in
`../assets/snippets.css`; fragments are fetched client-side by
`../assets/snippets.js` into any `.snippet-code[data-snippet="<name>"]`.

`pages/docs.html` was migrated from hand-escaped highlight.js blocks to this
system with a one-time script (`/tmp/migrate-docs.mjs`): it unescaped each
`<pre><code>` into `src/docs/docs-NN.<ext>` and replaced it with a placeholder.

## Regenerate

```sh
node snippets/build.mjs
```

Run it whenever a component or `src/` snippet changes. Requires the
`tree-sitter` CLI (>= 0.26).

## Layout

- `build.mjs` — the generator
- `src/` — standalone snippet sources (`src/docs/` = the docs code blocks)
- `.ts/config.json` — tree-sitter config (theme + parser dir)
- `.ts/grammars/` — cloned `tree-sitter-html` + `tree-sitter-go` (gitignored;
  re-clone with the two `git clone` lines below if missing)

```sh
git clone --depth 1 https://github.com/tree-sitter/tree-sitter-html snippets/.ts/grammars/tree-sitter-html
git clone --depth 1 https://github.com/tree-sitter/tree-sitter-go   snippets/.ts/grammars/tree-sitter-go
```

The HTML grammar's `queries/injections.scm` is emptied by `build.mjs`'s setup so
the HTML pass leaves the script body plain for splicing.
