#!/usr/bin/env node
/**
 * Set all OpenClaw cron jobs (agentTurn) to use a single model (default: cursor/auto).
 * Same as heartbeat/default — keeps cron and chat in sync.
 *
 * Run: node scripts/set-cron-model.js [--dry-run] [model]
 * Env: OPENCLAW_CRON_JOBS (optional path to jobs.json)
 *
 * Examples:
 *   node scripts/set-cron-model.js                    # set all to cursor/auto
 *   node scripts/set-cron-model.js cursor/opus-4.6    # set all to opus
 *   node scripts/set-cron-model.js --dry-run          # preview only
 *
 * On apply, creates a timestamped backup (jobs.json.bak.<ts>) before writing.
 */

const fs = require('fs');
const path = require('path');

const DEFAULT_MODEL = 'cursor/auto';
const jobsPath =
  process.env.OPENCLAW_CRON_JOBS ||
  path.join(process.env.HOME || process.env.USERPROFILE || '', '.openclaw', 'cron', 'jobs.json');

const argv = process.argv.slice(2);
const dryRun = argv.includes('--dry-run');
const model = argv.filter((a) => !a.startsWith('-'))[0] || DEFAULT_MODEL;

function main() {
  if (!fs.existsSync(jobsPath)) {
    console.error('Cron jobs file not found:', jobsPath);
    process.exit(1);
  }

  const raw = fs.readFileSync(jobsPath, 'utf8');
  let data;
  try {
    data = JSON.parse(raw);
  } catch (e) {
    console.error('Invalid JSON in', jobsPath, e.message);
    process.exit(1);
  }

  if (!data.jobs || !Array.isArray(data.jobs)) {
    console.error('Missing or invalid jobs array in', jobsPath);
    process.exit(1);
  }

  let updated = 0;
  for (const job of data.jobs) {
    const p = job.payload;
    if (!p || p.kind !== 'agentTurn') continue;

    const prev = p.model;
    if (prev === model) continue;

    p.model = model;
    updated++;
    console.log('Set model:', job.name, prev ? `(${prev} → ${model})` : `(→ ${model})`);
  }

  if (updated > 0) {
    if (dryRun) {
      console.log('[dry-run] Would update', updated, 'job(s). Run without --dry-run to apply.');
    } else {
      const backupPath = jobsPath + '.bak.' + Date.now();
      fs.writeFileSync(backupPath, raw);
      console.log('Backup written to', backupPath);
      fs.writeFileSync(jobsPath, JSON.stringify(data, null, 2) + '\n');
      console.log('Total updated:', updated, '→ model:', model);
    }
  } else {
    console.log('No jobs needed updates (all already', model + ').');
  }
}

main();
