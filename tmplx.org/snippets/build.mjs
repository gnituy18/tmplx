import { execFileSync } from 'node:child_process';
import { readFileSync, writeFileSync, readdirSync, mkdirSync, statSync } from 'node:fs';
import { join, basename, dirname, extname } from 'node:path';
import { fileURLToPath } from 'node:url';
import os from 'node:os';

const ROOT = dirname(fileURLToPath(import.meta.url));      // snippets/
const CONFIG = join(ROOT, '.ts/config.json');
const OUT = join(ROOT, '..', 'assets', 'snippets');
mkdirSync(OUT, { recursive: true });

// HTML pass must leave the tmplx <script> body plain so we can splice Go in.
writeFileSync(join(ROOT, '.ts/grammars/tree-sitter-html/queries/injections.scm'),
  '; managed by build.mjs - injections handled manually (two-pass splice)\n');

// Sources: real components (tmplx) + arbitrary doc snippets (.html/.go/.sh).
const SRC_ROOTS = [join(ROOT, '..', 'components'), join(ROOT, 'src')];

const escapeHtml = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

// tree-sitter highlight -> per-source-line HTML (inner of <td class=line>)
function hl(file) {
  const out = execFileSync('tree-sitter',
    ['highlight', '--config-path', CONFIG, '-H', '--css-classes', file], { encoding: 'utf8' });
  const rows = [];
  const re = /<td class=line>([\s\S]*?)<\/td>/g;
  let m;
  while ((m = re.exec(out))) rows.push(m[1].replace(/\n$/, ''));
  return rows;
}

function render(file, ext) {
  const raw = readFileSync(file, 'utf8').replace(/\n$/, '');
  if (ext === '.sh' || ext === '.txt') return escapeHtml(raw);   // plain
  if (ext === '.go') return hl(file).join('\n');                 // pure Go

  // .html -> HTML, with Go spliced into the tmplx <script> body if present
  const lines = raw.split('\n');
  const rows = hl(file);
  let a = -1, b = -1;
  for (let i = 0; i < lines.length; i++) {
    if (a < 0 && /<script[^>]*type=["']text\/tmplx["']/.test(lines[i])) a = i + 1;
    else if (a >= 0 && b < 0 && /<\/script>/.test(lines[i])) b = i - 1;
  }
  if (a >= 0 && b >= a) {
    const tmp = join(os.tmpdir(), `tmplx-snip-${basename(file, ext)}.go`);
    writeFileSync(tmp, lines.slice(a, b + 1).join('\n') + '\n');
    const go = hl(tmp);
    for (let i = a; i <= b; i++) if (go[i - a] !== undefined) rows[i] = go[i - a];
  }
  return rows.join('\n');
}

function walk(dir, out = []) {
  let entries;
  try { entries = readdirSync(dir); } catch { return out; }
  for (const e of entries) {
    if (e === '.ts') continue;
    const p = join(dir, e);
    if (statSync(p).isDirectory()) walk(p, out);
    else if (/\.(html|go|sh|txt)$/.test(e)) out.push(p);
  }
  return out;
}

let n = 0;
for (const root of SRC_ROOTS) {
  for (const file of walk(root).sort()) {
    const ext = extname(file);
    const name = basename(file, ext);
    writeFileSync(join(OUT, `${name}.html`),
      `<pre class="tmplx-code"><code>${render(file, ext)}</code></pre>\n`);
    n++;
  }
}
console.log(`built ${n} snippet fragments`);
