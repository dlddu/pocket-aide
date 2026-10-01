#!/usr/bin/env node
// 규칙 번호(`fail` 의 첫 인자)와 각 검사의 한계는 이 디렉터리 README.md 의 표가 정본이다 — 검사를 바꾸면 표도 같이 바꾼다.

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { JSDOM, VirtualConsole } from 'jsdom';

const here = path.dirname(fileURLToPath(import.meta.url));
const argRoot = process.argv.indexOf('--root');
const ROOT = path.resolve(argRoot > 0 ? process.argv[argRoot + 1] : path.join(here, '..', '..'));
const DOCS = path.join(ROOT, 'docs');
const JOURNEY_DOCS = path.join(DOCS, 'user-journeys');
const PAGES = path.join(DOCS, 'journeys');
const INDEX_MD = path.join(DOCS, 'mockups', '_index.md');
const HUB = path.join(DOCS, 'index.html');

const failures = [];
const fail = (rule, where, msg) => failures.push(`[${rule}] ${where}: ${msg}`);
let checks = 0;

const META_RES = [/\b(?:JRN|STP)-[a-z0-9]/, /\bPRD-\d/, /\bAC\s?\d/];
const INPUT_ROLES = new Set(['textbox', 'searchbox', 'combobox', 'listbox', 'checkbox', 'radio', 'switch', 'slider', 'spinbutton']);
const INPUT_LOOKALIKE = new Set(['field', 'input', 'textbox', 'textarea', 'select', 'pill', 'toggle', 'switch', 'checkbox', 'radio']);
const FORM_TAGS = ['INPUT', 'SELECT', 'TEXTAREA'];
const CLICKABLE = 'button, a[href], [data-go], input[type=submit], input[type=button], [role=button]';

function section(md, headingRe) {
  const lines = md.split('\n');
  const start = lines.findIndex((l) => headingRe.test(l));
  if (start < 0) return null;
  const level = lines[start].match(/^#+/)[0].length;
  const out = [];
  for (let i = start + 1; i < lines.length; i++) {
    const m = lines[i].match(/^(#+)\s/);
    if (m && m[1].length <= level) break;
    out.push(lines[i]);
  }
  return out.join('\n');
}

const tableRows = (text) => text.split('\n').filter((l) => /^\|/.test(l) && !/^\|\s*:?-/.test(l)).slice(1);

function readJourneys() {
  const out = new Map();
  for (const f of fs.readdirSync(JOURNEY_DOCS).filter((n) => /^JRN-.+\.md$/.test(n)).sort()) {
    const id = f.replace(/\.md$/, '');
    const md = fs.readFileSync(path.join(JOURNEY_DOCS, f), 'utf8');
    const status = (md.match(/^\|\s*상태\s*\|\s*([^|]+)\|/m) || [, ''])[1].trim();
    const steps = [...(section(md, /^##\s*3\.\s/) || '').matchAll(/^###\s+`(STP-[^`]+)`/gm)].map((m) => m[1]);
    const branchRows = tableRows(section(md, /^##\s*4\.\s/) || '').length;
    out.set(id, { id, status, steps, branchRows });
  }
  return out;
}

function readExceptions() {
  const readme = path.join(JOURNEY_DOCS, 'README.md');
  const sec = fs.existsSync(readme) ? section(fs.readFileSync(readme, 'utf8'), /^##\s*여정 mockup 예외\s*$/) : null;
  const rows = sec === null ? [] : tableRows(sec).filter((r) => /`JRN-[^`]+`/.test(r));
  return { ids: new Set(rows.map((r) => r.match(/`(JRN-[^`]+)`/)[1])), rows };
}

function open(file, hash) {
  const errors = [];
  const vc = new VirtualConsole();
  vc.on('jsdomError', (e) => errors.push(e.message));
  const dom = new JSDOM(fs.readFileSync(file, 'utf8'), {
    url: pathToFileURL(file).href + (hash ? '#' + hash : ''),
    runScripts: 'dangerously',
    pretendToBeVisual: true,
    virtualConsole: vc,
    beforeParse(w) { w.scrollTo = () => {}; },
  });
  return { w: dom.window, d: dom.window.document, errors };
}

const settle = () => new Promise((r) => setTimeout(r, 15));

async function load(file, hash) {
  const pg = open(file, hash);
  await settle();
  return pg;
}

function visible(w, el) {
  for (let n = el; n && n.nodeType === 1; n = n.parentElement) {
    if (n.hidden) return false;
    const cs = w.getComputedStyle(n);
    if (cs.display === 'none' || cs.visibility === 'hidden') return false;
    const p = n.parentElement;
    if (p && p.tagName === 'DETAILS' && !p.open && n.tagName !== 'SUMMARY') return false;
  }
  return true;
}

const isWrapper = (el) => !!el.closest('[class*="jm-"]');
const isBranchSelector = (el) => !!el.closest('.jm-branches');
const key = (step, state) => (state ? `${step}/${state}` : step);

// 지금 보이는 (단계, 상태)를 URL 이 아니라 화면에서 읽는다.
function where(pg) {
  const { w, d } = pg;
  const main = d.querySelector('main[data-journey]');
  const shown = [...main.querySelectorAll(':scope > section[data-step]')].filter((s) => visible(w, s));
  if (shown.length !== 1) return { error: `보이는 단계가 ${shown.length}개` };
  const sec = shown[0];
  const states = [...sec.querySelectorAll('[data-state]')].filter((s) => visible(w, s));
  if (states.length > 1) return { error: `${sec.dataset.step} 에 보이는 상태가 ${states.length}개` };
  const stateEl = states[0] || null;
  return {
    step: sec.dataset.step,
    state: stateEl ? stateEl.dataset.state : null,
    end: (stateEl && stateEl.getAttribute('data-end')) || sec.getAttribute('data-end'),
    sec,
  };
}

function actions(pg, at) {
  const inSec = [...at.sec.querySelectorAll(CLICKABLE)];
  const branches = [...pg.d.querySelectorAll(`.jm-branches :is(${CLICKABLE})`)].filter((el) => !at.sec.contains(el));
  return inSec.concat(branches).filter((el) => visible(pg.w, el) && !el.disabled && (!isWrapper(el) || isBranchSelector(el)));
}

function visibleText(w, root) {
  const out = [];
  const walk = (n) => {
    if (n.nodeType === 3) { if (n.data.trim()) out.push(n); return; }
    if (n.nodeType !== 1 || ['SCRIPT', 'STYLE', 'TEMPLATE'].includes(n.tagName)) return;
    for (const c of n.childNodes) walk(c);
  };
  walk(root);
  return out.filter((t) => visible(w, t.parentElement)).map((t) => t.data);
}

function checkMeta(pg, at_) {
  checks++;
  for (const t of visibleText(pg.w, pg.d.body)) {
    if (META_RES.some((re) => re.test(t))) { fail('5b', at_, `문서 메타가 펼쳐진 채 보인다: "${t.trim().slice(0, 60)}"`); break; }
  }
  for (const sel of ['.jm-steps', '.jm-pos']) {
    const el = pg.d.querySelector(sel);
    if (el && visible(pg.w, el) && el.textContent.trim()) fail('5b', at_, `${sel}(단계 번호·위치)가 펼쳐진 채 보인다`);
  }
}

function checkInputs(pg, at, at_) {
  const { w, d } = pg;
  checks++;
  const scope = [...at.sec.querySelectorAll('*')].filter((el) => !isWrapper(el) && visible(w, el));
  for (const el of scope) {
    const tag = el.tagName.toLowerCase();
    if (FORM_TAGS.includes(el.tagName)) continue;
    const ce = el.getAttribute('contenteditable');
    if (ce !== null && ce !== 'false') fail('5d', at_, `contenteditable <${tag} class="${el.className}"> — 실제 폼 요소가 아니다`);
    const role = el.getAttribute('role');
    if (role && INPUT_ROLES.has(role)) fail('5d', at_, `role="${role}" <${tag}> — 실제 폼 요소가 아니다`);
    if (el.hasAttribute('aria-pressed') || el.hasAttribute('aria-checked')) fail('5d', at_, `aria-pressed/checked 토글 <${tag}> — 실제 폼 요소가 아니다`);
    const look = [...el.classList].find((c) => INPUT_LOOKALIKE.has(c));
    const control = el.querySelector('input, select, textarea') || (tag === 'label' && el.control);
    if (look && !control) fail('5d', at_, `입력처럼 보이는 <${tag} class="${el.className}"> 에 실제 폼 요소가 없다`);
  }
  for (const el of scope.filter((e) => FORM_TAGS.includes(e.tagName))) {
    const name = `<${el.tagName.toLowerCase()}${el.type ? ` type=${el.type}` : ''}${el.name ? ` name=${el.name}` : ''}>`;
    if (el.disabled) { fail('5d', at_, `${name} 가 disabled`); continue; }
    el.focus();
    if (d.activeElement !== el) fail('5d', at_, `${name} 에 포커스가 가지 않는다`);
    const type = (el.type || '').toLowerCase();
    if (el.tagName === 'SELECT') {
      if (el.options.length < 2) { fail('5d', at_, `${name} 선택지가 ${el.options.length}개`); continue; }
      const want = (el.selectedIndex + 1) % el.options.length;
      el.selectedIndex = want;
      if (el.selectedIndex !== want) fail('5d', at_, `${name} 선택이 바뀌지 않는다`);
    } else if (type === 'radio' || type === 'checkbox') {
      const before = el.checked;
      el.click();
      if (type === 'radio' ? !el.checked : el.checked === before) fail('5d', at_, `${name} 를 눌러도 선택이 바뀌지 않는다`);
    } else if (!['submit', 'button', 'reset', 'hidden', 'image'].includes(type)) {
      if (el.readOnly) { fail('5d', at_, `${name} 가 readonly`); continue; }
      const v = el.value + ' 가';
      el.value = v;
      el.dispatchEvent(new w.Event('input', { bubbles: true }));
      if (el.value !== v) fail('5d', at_, `${name} 에 타이핑이 반영되지 않는다`);
    }
  }
}

// 행동마다 페이지를 새로 연다 — 앞 클릭이 바꾼 상태가 다음 간선 측정에 섞이지 않게.
async function outgoing(file, n, count) {
  const out = [];
  for (let i = 0; i < count; i++) {
    const pg = await load(file, n);
    const el = actions(pg, where(pg))[i];
    if (!el) continue;
    const screen = !isWrapper(el);
    el.click();
    await settle();
    const at = where(pg);
    if (!at.error) out.push({ to: key(at.step, at.state), screen });
  }
  return out;
}

async function checkPage(journey, file) {
  const rel = path.relative(ROOT, file);
  const base = await load(file);
  const { d } = base;
  for (const e of base.errors) fail('5h', rel, `스크립트 오류: ${e}`);

  const mains = d.querySelectorAll('main[data-journey]');
  checks++;
  if (mains.length !== 1) { fail('2', rel, `main[data-journey] 가 ${mains.length}개`); return null; }
  if (mains[0].dataset.journey !== journey.id) fail('2', rel, `선언한 여정 ${mains[0].dataset.journey} ≠ 디렉터리 ${journey.id}`);

  checks++;
  for (const s of d.querySelectorAll('script[src]')) {
    if (/^(https?:)?\/\//.test(s.getAttribute('src'))) fail('5h', rel, `외부 스크립트 ${s.getAttribute('src')}`);
  }
  for (const l of d.querySelectorAll('link[rel~=stylesheet][href]')) {
    const href = l.getAttribute('href');
    if (/^(https?:)?\/\//.test(href) && !/fonts\.(googleapis|gstatic)\.com/.test(href)) fail('5h', rel, `외부 스타일시트 ${href}`);
  }

  const sections = [...mains[0].querySelectorAll(':scope > section[data-step]')];
  const pageSteps = sections.map((s) => s.dataset.step);
  checks++;
  const dup = pageSteps.filter((s, i) => pageSteps.indexOf(s) !== i);
  if (dup.length) fail('3', rel, `중복 단계 ${dup.join(', ')}`);
  for (const s of journey.steps) if (!pageSteps.includes(s)) fail('3', rel, `문서에만 있는 단계 ${s}`);
  for (const s of pageSteps) if (!journey.steps.includes(s)) fail('3', rel, `페이지에만 있는 단계 ${s}`);
  if (!pageSteps.length) return null;

  const states = new Map(sections.map((s) => [s.dataset.step, [...s.querySelectorAll('[data-state]')].map((e) => e.dataset.state)]));
  const nodes = pageSteps.flatMap((s) => [key(s, null), ...states.get(s).map((st) => key(s, st))]);

  checks++;
  for (const el of d.querySelectorAll('[data-go]')) {
    const go = el.getAttribute('data-go');
    const [s, st] = go.split('/');
    if (!states.has(s)) fail('6', rel, `data-go="${go}" — 없는 단계`);
    else if (st && !states.get(s).includes(st)) fail('6', rel, `data-go="${go}" — ${s} 에 없는 상태`);
  }

  checks++;
  const stateCount = nodes.length - pageSteps.length;
  if (stateCount < journey.branchRows) fail('5e', rel, `분기·예외 ${journey.branchRows}행 > 상태 ${stateCount}개`);

  const edges = new Map();
  const ends = new Map();
  for (const n of nodes) {
    const [step, state] = n.split('/');
    const pg = await load(file, n);
    const at = where(pg);
    checks++;
    if (at.error) { fail('5f', `${rel}#${n}`, at.error); continue; }
    if (at.step !== step || at.state !== (state || null)) { fail('5f', `${rel}#${n}`, `딥링크가 ${key(at.step, at.state)} 을 연다`); continue; }
    for (const e of pg.errors) fail('5h', `${rel}#${n}`, `스크립트 오류: ${e}`);
    checkMeta(pg, `${rel}#${n}`);
    checkInputs(pg, at, `${rel}#${n}`);
    ends.set(n, at.end);
    edges.set(n, await outgoing(file, n, actions(pg, at).length));
  }

  for (let i = 0; i < pageSteps.length - 1; i++) {
    checks++;
    if (!(edges.get(pageSteps[i]) || []).some((e) => e.screen && e.to === pageSteps[i + 1])) {
      fail('5c', `${rel}#${pageSteps[i]}`, `화면 안 행동으로 ${pageSteps[i + 1]} 에 닿지 않는다`);
    }
  }

  const seen = new Set([pageSteps[0]]);
  const queue = [pageSteps[0]];
  while (queue.length) {
    for (const e of edges.get(queue.shift()) || []) if (!seen.has(e.to)) { seen.add(e.to); queue.push(e.to); }
  }
  for (const n of nodes) {
    checks++;
    if (!seen.has(n)) fail(n.includes('/') ? '5e' : '5c', `${rel}#${n}`, '첫 단계에서 화면 행동·분기 선택으로 도달할 수 없다');
  }

  for (const n of nodes) {
    checks++;
    const out = (edges.get(n) || []).filter((e) => e.screen && e.to !== n);
    if (edges.has(n) && !out.length && !ends.get(n)) fail('5g', `${rel}#${n}`, '화면 안에 나가는 행동이 없는데 data-end 로 끝을 표시하지 않았다');
  }
  checks++;
  if (!sections[sections.length - 1].getAttribute('data-end')) fail('5g', `${rel}#${pageSteps[pageSteps.length - 1]}`, '마지막 단계에 data-end 가 없다');

  return { pageSteps, states };
}

function checkIndexes(pages) {
  const idx = fs.existsSync(INDEX_MD) ? section(fs.readFileSync(INDEX_MD, 'utf8'), /^##\s*여정 mockup\s*$/) : null;
  const hub = fs.existsSync(HUB) ? fs.readFileSync(HUB, 'utf8') : '';
  checks++;
  if (idx === null) { fail('7', 'docs/mockups/_index.md', '「여정 mockup」 절이 없다'); return; }
  const listed = new Set([...idx.matchAll(/^###\s+journeys\/(JRN-[^/\s]+)\//gm)].map((m) => m[1]));
  const linked = new Set([...hub.matchAll(/href="journeys\/(JRN-[^/"]+)\/"/g)].map((m) => m[1]));
  for (const [id, info] of pages) {
    checks++;
    if (!listed.has(id)) fail('7', 'docs/mockups/_index.md', `「여정 mockup」 절에 journeys/${id}/ 가 없다`);
    else {
      const sub = section(idx, new RegExp(`^###\\s+journeys/${id}/`));
      for (const s of info.pageSteps) if (!sub.includes('`' + s + '`')) fail('7', 'docs/mockups/_index.md', `${id} 의 담은 단계에 ${s} 가 없다`);
      const want = new Set([...info.states].flatMap(([s, sts]) => sts.map((st) => `${s}/${st}`)));
      const got = new Set([...sub.matchAll(/`(STP-[^`/]+\/[^`]+)`/g)].map((m) => m[1]));
      for (const k of want) if (!got.has(k)) fail('7', 'docs/mockups/_index.md', `${id} 의 분기 상태에 ${k} 가 없다`);
      for (const k of got) if (!want.has(k)) fail('7', 'docs/mockups/_index.md', `${id} 의 분기 상태 ${k} 가 페이지에 없다`);
    }
    if (!linked.has(id)) fail('7', 'docs/index.html', `journeys/${id}/ 링크가 없다`);
  }
  for (const id of listed) if (!pages.has(id)) fail('7', 'docs/mockups/_index.md', `없는 페이지 journeys/${id}/ 를 매핑한다`);
  for (const id of linked) if (!pages.has(id)) fail('7', 'docs/index.html', `없는 페이지 journeys/${id}/ 로 링크한다`);
}

const journeys = readJourneys();
const exceptions = readExceptions();
const pageDirs = fs.existsSync(PAGES)
  ? fs.readdirSync(PAGES).filter((n) => fs.existsSync(path.join(PAGES, n, 'index.html'))).sort()
  : [];

for (const r of exceptions.rows) {
  checks++;
  const id = r.match(/`(JRN-[^`]+)`/)[1];
  if (!journeys.has(id)) fail('6', 'docs/user-journeys/README.md', `예외 목록이 없는 여정 ${id} 을 등재한다`);
  const cells = r.split('|').slice(1, -1).map((c) => c.trim());
  if (cells.length < 3 || cells.some((c) => !c)) fail('8', 'docs/user-journeys/README.md', `${id} 예외 행에 사유·재검토 시점이 비어 있다`);
}

const judged = [...journeys.values()].filter((j) => !/폐기/.test(j.status) && !(exceptions.ids.has(j.id) && /초안/.test(j.status)));
for (const j of judged) {
  checks++;
  if (!pageDirs.includes(j.id) && !exceptions.ids.has(j.id)) fail('1', `docs/user-journeys/${j.id}.md`, '여정 mockup 페이지가 없다(예외 미등재)');
}
for (const dir of pageDirs) {
  checks++;
  if (!journeys.has(dir)) fail('2', `docs/journeys/${dir}/index.html`, '대응하는 여정 문서가 없다(고아 페이지)');
  else if (/폐기/.test(journeys.get(dir).status)) fail('6', `docs/journeys/${dir}/index.html`, '폐기된 여정의 페이지다');
}

const pages = new Map();
for (const dir of pageDirs) {
  if (!journeys.has(dir)) continue;
  const info = await checkPage(journeys.get(dir), path.join(PAGES, dir, 'index.html'));
  if (info) pages.set(dir, info);
}
checkIndexes(pages);

console.log(`journey-mockup-harness: 여정 ${journeys.size} · 판정 대상 ${judged.length} · 예외 ${exceptions.ids.size} · 페이지 ${pageDirs.length} · 검사 ${checks} · 위반 ${failures.length}`);
for (const f of failures) console.log('  ✗ ' + f);
process.exit(failures.length ? 1 : 0);
