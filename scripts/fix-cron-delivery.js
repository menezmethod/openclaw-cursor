#!/usr/bin/env node
/**
 * Fix OpenClaw cron job delivery so all announce jobs have explicit
 * channel + to, avoiding WhatsApp fallback errors.
 *
 * Run: node scripts/fix-cron-delivery.js [--dry-run]
 * Env: OPENCLAW_CRON_JOBS (optional path to jobs.json)
 *
 * On apply, creates a timestamped backup (jobs.json.bak.<ts>) before writing.
 * Restore: cp ~/.openclaw/cron/jobs.json.bak.<ts> ~/.openclaw/cron/jobs.json
 */

const fs = require('fs');
const path = require('path');

// Set OPENCLAW_CRON_DEFAULT_TO to override (default: User's Telegram ID)
const DEFAULT_TELEGRAM_TO = process.env.OPENCLAW_CRON_DEFAULT_TO || 'YOUR_TELEGRAM_ID';
const jobsPath =
  process.env.OPENCLAW_CRON_JOBS ||
  path.join(process.env.HOME || process.env.USERPROFILE || '', '.openclaw', 'cron', 'jobs.json');
const dryRun = process.argv.includes('--dry-run');

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
    const d = job.delivery;
    if (!d) continue;

    const mode = d.mode || 'announce';
    let changed = false;
    let next = { ...d, mode };

    // "to": "User" with no channel → use Telegram + numeric ID
    if (d.to === 'User' && !d.channel) {
      next.channel = 'telegram';
      next.to = DEFAULT_TELEGRAM_TO;
      changed = true;
    }

    // channel is telegram but no to → add to so delivery doesn't fail
    if (d.channel === 'telegram' && !d.to) {
      next.to = DEFAULT_TELEGRAM_TO;
      changed = true;
    }

    // Jobs with explicit non-User target (e.g. to: "webchat") are left unchanged — we only set channel/to when missing or "User"

    // mode "announce" with no channel and no to → default to Telegram (avoids WhatsApp fallback)
    if (mode === 'announce' && !d.channel && !d.to) {
      next.channel = 'telegram';
      next.to = DEFAULT_TELEGRAM_TO;
      changed = true;
    }

    if (changed) {
      job.delivery = next;
      updated++;
      console.log('Fixed:', job.name);
    }
  }

  if (updated > 0) {
    if (dryRun) {
      console.log('[dry-run] Would update', updated, 'job(s). Run without --dry-run to apply.');
    } else {
      const backupPath = jobsPath + '.bak.' + Date.now();
      fs.writeFileSync(backupPath, raw);
      console.log('Backup written to', backupPath);
      fs.writeFileSync(jobsPath, JSON.stringify(data, null, 2) + '\n');
      console.log('Total updated:', updated);
    }
  } else {
    console.log('No jobs needed updates.');
  }
}

main();
