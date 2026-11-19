import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { createRequire } from 'node:module';

// Use Vite's own CSS parser so tests validate rules, not textual snapshots.
const require = createRequire(import.meta.url);
const postcss = createRequire(require.resolve('vite/package.json'))('postcss');
const base = postcss.parse(readFileSync(new URL('../src/styles.css', import.meta.url), 'utf8'));
const workspace = postcss.parse(readFileSync(new URL('../src/workspace-overrides.css', import.meta.url), 'utf8'));
function tokens(selector) {
  const result = {};
  base.walkRules(selector, rule => rule.walkDecls(/^--/, decl => result[decl.prop] = decl.value));
  return result;
}
const light = tokens(':root');
const dark = tokens(':root[data-theme="dark"]');
const system = tokens(':root[data-theme="system"]');
const themeNames = ['status-text','column-bg','column-border','column-heading','priority-high','priority-critical','card-shadow','detail-muted','description-text','code-bg','timeline-bg','timeline-border'];

test('explicit and system dark themes cover every workspace semantic token', () => {
  for (const name of themeNames) {
    assert.ok(light[`--${name}`], `light ${name}`);
    assert.ok(dark[`--${name}`], `dark ${name}`);
    assert.equal(system[`--${name}`], dark[`--${name}`], `system ${name}`);
    assert.notEqual(dark[`--${name}`], light[`--${name}`], `${name} must adapt`);
  }
  base.walkRules(':root[data-theme="system"]', rule => {
    assert.equal(rule.parent.name, 'media');
    assert.equal(rule.parent.params, '(prefers-color-scheme:dark)');
  });
});

test('light variant A retains its column and priority palette', () => {
  assert.equal(light['--column-bg'], '#eceeea');
  assert.equal(light['--column-border'], '#d7dbd6');
  assert.equal(light['--column-heading'], '#414950');
  assert.equal(light['--priority-high'], '#334f82');
  assert.equal(light['--priority-critical'], '#8b3e52');
});

function luminance(hex) {
  if (hex.length === 4) hex = '#' + [...hex.slice(1)].map(c => c+c).join('');
  const rgb = hex.slice(1).match(/../g).map(v => parseInt(v,16)/255).map(v => v <= .04045 ? v/12.92 : ((v+.055)/1.055)**2.4);
  return rgb[0]*.2126 + rgb[1]*.7152 + rgb[2]*.0722;
}
function contrast(a,b) {const x=luminance(a), y=luminance(b);return (Math.max(x,y)+.05)/(Math.min(x,y)+.05);}
test('workspace labels and priority text retain readable contrast in both themes', () => {
  for (const theme of [light,dark]) {
    assert.ok(contrast(theme['--column-heading'],theme['--column-bg']) >= 4.5);
    for (const name of ['priority-high','priority-critical','detail-muted','description-text','status-text']) {
      assert.ok(contrast(theme[`--${name}`],theme['--surface']) >= 4.5, name);
    }
  }
});

test('workspace columns, labels, code and timeline consume theme tokens rather than fixed light colors', () => {
  for (const selector of ['.kanban-column','.kanban-column .column-heading','.task-card','.task-card .chips .priority.high','.task-card .chips .priority.critical']) {
    let found = false;
    workspace.walkRules(selector, rule => {
      found = true;
      rule.walkDecls(decl => assert.doesNotMatch(decl.value, /#[a-f\d]{3,8}\b/i, `${selector} ${decl.prop}`));
    });
    assert.ok(found, selector);
  }
  for (const selector of ['.markdown code','.timeline article','.timeline small','.task-fields dt']) {
    base.walkRules(selector, rule => rule.walkDecls(decl => assert.doesNotMatch(decl.value, /#[a-f\d]{3,8}\b/i, selector)));
  }
});
