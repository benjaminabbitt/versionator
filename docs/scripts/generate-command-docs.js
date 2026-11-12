#!/usr/bin/env node
/**
 * Schema-driven documentation generator for versionator
 *
 * Reads the CLI schema from `versionator schema` and generates:
 * - Command reference pages (docs/commands/*.md)
 * - Template variables reference (docs/templates/variables.md)
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const DOCS_DIR = path.join(__dirname, '..', 'docs');
const COMMANDS_DIR = path.join(DOCS_DIR, 'commands');
const TEMPLATES_DIR = path.join(DOCS_DIR, 'templates');

// Generate schema from versionator
function getSchema() {
  try {
    const output = execSync('versionator support schema', { encoding: 'utf-8' });
    return JSON.parse(output);
  } catch (error) {
    console.error('Error running versionator support schema:', error.message);
    console.error('Make sure versionator is installed and in your PATH');
    process.exit(1);
  }
}

// Escape MDX special characters
function escapeMdx(str) {
  if (!str) return str;
  // Escape < and > to prevent MDX from parsing them as tags
  // Escape {{ and }} to prevent MDX from parsing them as JSX expressions
  return str
    .replace(/</g, '\\<')
    .replace(/>/g, '\\>')
    .replace(/\{\{/g, '\\{\\{')
    .replace(/\}\}/g, '\\}\\}');
}

// Format long description, converting Examples: sections to proper code blocks
function formatLongDescription(text) {
  if (!text) return '';

  // Split on "Examples:" to separate description from examples
  const parts = text.split(/\n\nExamples:\n/);
  if (parts.length === 1) {
    // No examples section, check if it has template vars
    if (text.includes('{{')) {
      return '```\n' + text + '\n```\n\n';
    }
    return escapeMdx(text) + '\n\n';
  }

  // Has examples section
  let result = '';
  const description = parts[0];
  const examples = parts[1];

  // Format description
  if (description.includes('{{')) {
    result += '```\n' + description + '\n```\n\n';
  } else {
    result += escapeMdx(description) + '\n\n';
  }

  // Format examples as code block
  result += '**Examples:**\n\n```bash\n';
  // Clean up the examples - remove leading spaces
  const cleanedExamples = examples.split('\n')
    .map(line => line.replace(/^  /, ''))
    .join('\n')
    .trim();
  result += cleanedExamples + '\n```\n\n';

  return result;
}

// Generate frontmatter for Docusaurus
function frontmatter(title, description, sidebar_position) {
  let fm = `---
title: ${title}
description: ${description}
`;
  if (sidebar_position !== undefined) {
    fm += `sidebar_position: ${sidebar_position}\n`;
  }
  fm += '---\n\n';
  return fm;
}

// Generate markdown for a single command
function generateCommandDoc(cmd, parentPath = []) {
  const fullPath = [...parentPath, cmd.name].join(' ');
  const filename = cmd.name + '.md';

  let md = frontmatter(
    `${cmd.name}`,
    cmd.short || `The ${cmd.name} command`
  );

  // Command header
  md += `# ${cmd.name}\n\n`;
  md += `${escapeMdx(cmd.short)}\n\n`;

  if (cmd.long) {
    md += formatLongDescription(cmd.long);
  }

  // Usage
  md += `## Usage\n\n`;
  md += '```bash\n';
  md += `versionator ${fullPath}`;
  if (cmd.subcommands && cmd.subcommands.length > 0) {
    md += ' [command]';
  }
  if (cmd.flags && cmd.flags.length > 0) {
    md += ' [flags]';
  }
  md += '\n```\n\n';

  // Aliases
  if (cmd.aliases && cmd.aliases.length > 0) {
    md += `**Aliases:** \`${cmd.aliases.join('`, `')}\`\n\n`;
  }

  // Subcommands
  if (cmd.subcommands && cmd.subcommands.length > 0) {
    md += `## Subcommands\n\n`;
    md += '| Command | Description |\n';
    md += '|---------|-------------|\n';
    for (const sub of cmd.subcommands) {
      md += `| \`${sub.name}\` | ${sub.short} |\n`;
    }
    md += '\n';

    // Document subcommands inline
    for (const sub of cmd.subcommands) {
      md += `### ${sub.name}\n\n`;
      md += `${escapeMdx(sub.short)}\n\n`;
      if (sub.long && sub.long !== sub.short) {
        md += formatLongDescription(sub.long);
      }
      md += '```bash\n';
      md += `versionator ${fullPath} ${sub.name}`;
      if (sub.flags && sub.flags.length > 0) {
        md += ' [flags]';
      }
      md += '\n```\n\n';

      if (sub.aliases && sub.aliases.length > 0) {
        md += `**Aliases:** \`${sub.aliases.join('`, `')}\`\n\n`;
      }

      if (sub.flags && sub.flags.length > 0) {
        md += '**Flags:**\n\n';
        md += '| Flag | Type | Default | Description |\n';
        md += '|------|------|---------|-------------|\n';
        for (const flag of sub.flags) {
          const flagName = flag.shorthand
            ? `-${flag.shorthand}, --${flag.name}`
            : `--${flag.name}`;
          const defaultVal = escapeMdx(flag.default) || '-';
          md += `| \`${flagName}\` | ${flag.type} | ${defaultVal} | ${escapeMdx(flag.description)} |\n`;
        }
        md += '\n';
      }
    }
  }

  // Flags
  if (cmd.flags && cmd.flags.length > 0) {
    md += `## Flags\n\n`;
    md += '| Flag | Type | Default | Description |\n';
    md += '|------|------|---------|-------------|\n';
    for (const flag of cmd.flags) {
      const flagName = flag.shorthand
        ? `-${flag.shorthand}, --${flag.name}`
        : `--${flag.name}`;
      const defaultVal = escapeMdx(flag.default) || '-';
      md += `| \`${flagName}\` | ${flag.type} | ${defaultVal} | ${escapeMdx(flag.description)} |\n`;
    }
    md += '\n';
  }

  return { filename, content: md };
}

// Generate template variables documentation
function generateVariablesDoc(templateVars) {
  let md = frontmatter(
    'Template Variables',
    'Complete reference of all template variables available in versionator',
    1
  );

  md += '# Template Variables\n\n';
  md += 'Versionator uses [Mustache](https://mustache.github.io/) templating. ';
  md += 'Use `{{VariableName}}` syntax in templates.\n\n';
  md += ':::tip\n';
  md += 'Run `versionator vars` to see all variables with their current values.\n';
  md += ':::\n\n';

  const categories = [
    { key: 'versionComponents', title: 'Version Components', description: 'Core version numbers and formatting' },
    { key: 'preRelease', title: 'Pre-release', description: 'Pre-release identifier variables' },
    { key: 'metadata', title: 'Build Metadata', description: 'Build metadata variables' },
    {
      key: 'vcs', title: 'VCS / Git Information', description: 'Version control information',
      notes: '`{{UncommittedChanges}}`, `{{Dirty}}`, `{{DateTimeDirty}}` and the `ci` `DIRTY` ' +
        'variable read one count, so they always agree. By default untracked files count. ' +
        'With `dirty.ignoreUntracked: true` in `.versionator.yaml` (or the global ' +
        '`--ignore-untracked` flag, which overrides it) only staged or unstaged changes to ' +
        'tracked files count, as `git describe --dirty` judges a tree: a checkout whose only ' +
        "change is an untracked file is clean. The `release` command's clean-tree check " +
        'follows the same setting.',
    },
    { key: 'commitInfo', title: 'Commit Information', description: 'Details about the current commit' },
    { key: 'buildTimestamps', title: 'Build Timestamps', description: 'Timestamps at build time' },
    {
      key: 'release', title: 'Release Templates',
      description: 'Available to the release tag, branch and message templates ' +
        '(`release.tagTemplate`, `release.branchTemplate`, `release.messageTemplate`) and to the ' +
        '`updates:` templates of every version change (`{{Tag}}` excepted, which exists only once ' +
        'release has rendered the tag). In all of them `{{Prefix}}` is the effective prefix: the ' +
        "VERSION file's, else `v`, as the default tag has it (release's `--prefix` flag replaces it there)",
    },
  ];

  for (const cat of categories) {
    const vars = templateVars[cat.key];
    if (!vars || vars.length === 0) continue;

    md += `## ${cat.title}\n\n`;
    md += `${cat.description}.\n\n`;
    md += '| Variable | Description | Example |\n';
    md += '|----------|-------------|--------|\n';
    for (const v of vars) {
      const example = v.example ? `\`${v.example}\`` : '-';
      md += `| \`{{${v.name}}}\` | ${v.description} | ${example} |\n`;
    }
    md += '\n';
    if (cat.notes) {
      md += `${cat.notes}\n\n`;
    }
  }

  return md;
}

// Generate commands index page
function generateCommandsIndex(commands, globalFlags) {
  let md = frontmatter(
    'Commands Reference',
    'Complete reference of all versionator commands',
    0
  );

  md += '# Commands Reference\n\n';
  md += 'Versionator provides commands for managing semantic versions.\n\n';

  md += '## Available Commands\n\n';
  md += '| Command | Description |\n';
  md += '|---------|-------------|\n';
  for (const cmd of commands) {
    md += `| [\`${cmd.name}\`](./${cmd.name}) | ${cmd.short} |\n`;
  }
  md += '\n';

  md += '## Global Flags\n\n';
  md += 'These flags are available on all commands:\n\n';
  md += '| Flag | Description |\n';
  md += '|------|-------------|\n';
  // From the schema, so a new global flag is listed without editing this file.
  for (const flag of globalFlags || []) {
    md += `| \`--${flag.name}\` | ${flag.description} |\n`;
  }
  md += '| `-h, --help` | Help for any command |\n';

  return md;
}

// Main execution
function main() {
  console.log('Generating documentation from versionator schema...\n');

  // Get schema
  const schema = getSchema();
  console.log(`Schema version: ${schema.version}`);
  console.log(`Commands found: ${schema.commands.length}`);

  // Ensure directories exist
  fs.mkdirSync(COMMANDS_DIR, { recursive: true });
  fs.mkdirSync(TEMPLATES_DIR, { recursive: true });

  // Generate commands index
  const indexContent = generateCommandsIndex(schema.commands, schema.globalFlags);
  const indexPath = path.join(COMMANDS_DIR, 'index.md');
  fs.writeFileSync(indexPath, indexContent);
  console.log(`Generated: ${indexPath}`);

  // Generate individual command docs
  for (const cmd of schema.commands) {
    const { filename, content } = generateCommandDoc(cmd);
    const filePath = path.join(COMMANDS_DIR, filename);
    fs.writeFileSync(filePath, content);
    console.log(`Generated: ${filePath}`);
  }

  // Generate template variables doc
  const varsContent = generateVariablesDoc(schema.templateVariables);
  const varsPath = path.join(TEMPLATES_DIR, 'variables.md');
  fs.writeFileSync(varsPath, varsContent);
  console.log(`Generated: ${varsPath}`);

  console.log('\nDone! Generated documentation from schema.');
}

main();
