#!/usr/bin/env node
// Create a GitHub App from a manifest in .github/app-manifests/ using GitHub's
// App Manifest flow, which is the only way to create an App with a preset
// permission list. The REST API cannot create Apps from a PAT, so this opens a
// browser, you click "Create GitHub App" once, and the redirect is captured here.
//
//   node scripts/create-app.mjs archilles-engineer
//   node scripts/create-app.mjs archilles-engineer --org my-org
//
// Ownership matters and is not changeable afterwards: an App marked private
// can only be installed on the account that owns it. If the repo is going to
// live in an org, create the Apps under that org (--org) or you will recreate
// all five after the transfer.
//
// Credentials land in .secrets/<app>/ (gitignored, mode 0600). Nothing is
// printed to stdout except the App id and slug.

import { createServer } from 'node:http';
import { readFileSync, mkdirSync, writeFileSync, chmodSync } from 'node:fs';
import { randomBytes } from 'node:crypto';
import { spawn } from 'node:child_process';

const app = process.argv[2];
if (!app) {
  console.error('usage: node scripts/create-app.mjs <app-name>');
  console.error('apps: archilles-engineer archilles-architect archilles-orchestrator archilles-bot archilles-release');
  process.exit(1);
}

const PORT = Number(process.env.PORT || 8787);
const manifest = JSON.parse(readFileSync(`.github/app-manifests/${app}.json`, 'utf8'));
delete manifest._comment;
// localhost is fine for redirect_url - GitHub explicitly supports it for this
// flow - but a hook url must be publicly reachable and is validated even when
// active:false. These Apps are agent-driven and poll, so they get no webhook.
manifest.redirect_url = `http://localhost:${PORT}/callback`;
manifest.hook_attributes = { active: false };
delete manifest.default_events;

const orgArg = process.argv[process.argv.indexOf('--org') + 1];
const org = (process.argv.includes('--org') && orgArg) || process.env.APP_ORG || null;
const newAppUrl = org
  ? `https://github.com/organizations/${org}/settings/apps/new`
  : 'https://github.com/settings/apps/new';

const state = randomBytes(16).toString('hex');
const perms = Object.entries(manifest.default_permissions).map(([k, v]) => `${k}:${v}`).join(', ');
const page = `<!doctype html><meta charset="utf-8"><title>Create ${app}</title>
<body style="font:14px system-ui;padding:3rem;max-width:40rem;margin:auto">
<h2>Create GitHub App: <code>${app}</code></h2>
<p>${manifest.description}</p>
<p><b>Owner:</b> ${org ? `organization <code>${org}</code>` : 'your personal account'}</p>
<p><b>Permissions:</b> ${perms}</p>
<form id=f method=post action="${newAppUrl}?state=${state}">
  <input type=hidden name=manifest value='${JSON.stringify(manifest).replace(/'/g, '&apos;')}'>
  <button type=submit style="font-size:1rem;padding:.6rem 1.2rem">Continue to GitHub &rarr;</button>
</form>
<script>document.getElementById('f').submit()</script>`;

const server = createServer(async (req, res) => {
  const url = new URL(req.url, `http://localhost:${PORT}`);

  if (url.pathname === '/') {
    res.writeHead(200, { 'content-type': 'text/html' });
    return res.end(page);
  }

  if (url.pathname === '/callback') {
    const code = url.searchParams.get('code');
    if (url.searchParams.get('state') !== state || !code) {
      res.writeHead(400, { 'content-type': 'text/html' });
      res.end('<h2>State mismatch or missing code. Re-run the script.</h2>');
      return server.close(() => process.exit(1));
    }
    const r = await fetch(`https://api.github.com/app-manifests/${code}/conversions`, {
      method: 'POST',
      headers: { accept: 'application/vnd.github+json', 'user-agent': 'archilles-setup' },
    });
    const body = await r.json();
    if (!r.ok) {
      res.writeHead(500, { 'content-type': 'text/html' });
      res.end(`<h2>Conversion failed</h2><pre>${JSON.stringify(body, null, 2)}</pre>`);
      return server.close(() => process.exit(1));
    }

    const dir = `.secrets/${app}`;
    mkdirSync(dir, { recursive: true, mode: 0o700 });
    const write = (name, data) => {
      const p = `${dir}/${name}`;
      writeFileSync(p, data);
      chmodSync(p, 0o600);
    };
    write('private-key.pem', body.pem);
    write('app.json', JSON.stringify(
      { id: body.id, slug: body.slug, node_id: body.node_id, client_id: body.client_id, owner: body.owner?.login },
      null, 2));
    write('client-secret.txt', body.client_secret ?? '');
    write('webhook-secret.txt', body.webhook_secret ?? '');

    const varName = app.toUpperCase().replace(/-/g, '_') + '_APP_ID';
    res.writeHead(200, { 'content-type': 'text/html' });
    res.end(`<body style="font:14px system-ui;padding:3rem">
      <h2>Created <code>${body.slug}</code> (id ${body.id})</h2>
      <p>Credentials written to <code>${dir}/</code> (gitignored, mode 0600).</p>
      <p><b>Next:</b> <a href="https://github.com/settings/apps/${body.slug}/installations">install it on jameslett/archilles</a>,
      then record the App id as a repo variable:</p>
      <pre>gh variable set ${varName} --body ${body.id} --repo jameslett/archilles</pre>
      <p>You can close this tab.</p>`);
    console.log(`${body.slug}\tid=${body.id}\tcreds=${dir}/`);
    return server.close(() => process.exit(0));
  }

  res.writeHead(404).end();
});

server.listen(PORT, () => {
  const target = `http://localhost:${PORT}/`;
  console.error(`Opening ${target} - click "Create GitHub App" in the browser.`);
  spawn(process.platform === 'darwin' ? 'open' : 'xdg-open', [target], { stdio: 'ignore', detached: true }).unref();
});
