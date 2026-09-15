#!/usr/bin/env node
// Mint a short-lived GitHub App installation token so an agent can act AS the
// App - commit, push, open PRs, review - rather than as the human running it.
// This is the local equivalent of what Renovate does in CI.
//
//   node scripts/app-token.mjs archilles-engineer              # print token
//   eval "$(node scripts/app-token.mjs archilles-engineer --env)"   # export GH_TOKEN
//
// Reads .secrets/<app>/{app.json,private-key.pem} written by create-app.mjs,
// or APP_ID / APP_PRIVATE_KEY from the environment when running somewhere the
// key lives in a secret store instead of on disk.
//
// Tokens expire in 1 hour and are scoped to this repository only.

import { readFileSync, existsSync } from 'node:fs';
import { createSign } from 'node:crypto';

const app = process.argv[2];
const flags = new Set(process.argv.slice(3));
if (!app) {
  console.error('usage: node scripts/app-token.mjs <app-name> [--env] [--repo owner/name]');
  process.exit(1);
}

const repoArg = process.argv[process.argv.indexOf('--repo') + 1];
const REPO = (process.argv.includes('--repo') && repoArg) || process.env.REPO || 'graphward/archilles';
const [owner, repoName] = REPO.split('/');

const dir = `.secrets/${app}`;
let appId = process.env.APP_ID;
let pem = process.env.APP_PRIVATE_KEY;
if (!appId || !pem) {
  if (!existsSync(`${dir}/app.json`)) {
    console.error(`No credentials for ${app}. Run: node scripts/create-app.mjs ${app}`);
    process.exit(1);
  }
  appId = String(JSON.parse(readFileSync(`${dir}/app.json`, 'utf8')).id);
  pem = readFileSync(`${dir}/private-key.pem`, 'utf8');
}

const b64 = (o) => Buffer.from(typeof o === 'string' ? o : JSON.stringify(o))
  .toString('base64url');

function jwt() {
  const now = Math.floor(Date.now() / 1000);
  // iat backdated 60s to tolerate clock skew. GitHub rejects exp - iat > 600,
  // so this leaves a full minute of headroom rather than sitting on the limit.
  const signingInput = `${b64({ alg: 'RS256', typ: 'JWT' })}.${b64({ iat: now - 60, exp: now + 480, iss: appId })}`;
  const sig = createSign('RSA-SHA256').update(signingInput).end().sign(pem).toString('base64url');
  return `${signingInput}.${sig}`;
}

async function gh(path, token, init = {}) {
  const r = await fetch(`https://api.github.com${path}`, {
    ...init,
    headers: {
      authorization: `Bearer ${token}`,
      accept: 'application/vnd.github+json',
      'x-github-api-version': '2022-11-28',
      'user-agent': 'archilles-agent',
      ...(init.headers || {}),
    },
  });
  const body = await r.json().catch(() => ({}));
  if (!r.ok) {
    console.error(`GitHub ${r.status} on ${path}: ${body.message || JSON.stringify(body)}`);
    if (r.status === 404 && path.startsWith('/repos')) {
      console.error(`Is ${app} installed on ${REPO}? https://github.com/settings/apps/${app}/installations`);
    }
    process.exit(1);
  }
  return body;
}

const appJwt = jwt();
const installation = await gh(`/repos/${owner}/${repoName}/installation`, appJwt);
const { token, expires_at } = await gh(
  `/app/installations/${installation.id}/access_tokens`,
  appJwt,
  { method: 'POST', body: JSON.stringify({ repositories: [repoName] }) },
);

if (flags.has('--env')) {
  // Consumed via eval; the token is the whole point so it goes to stdout.
  process.stdout.write(
    `export GH_TOKEN=${token}\n` +
    `export GIT_ASKPASS=\n` +
    `export ARCHILLES_ACTING_AS=${app}\n` +
    `export ARCHILLES_TOKEN_EXPIRES=${expires_at}\n`,
  );
} else {
  process.stdout.write(token + '\n');
}
console.error(`acting as ${app} on ${REPO}, expires ${expires_at}`);
