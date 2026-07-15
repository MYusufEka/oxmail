export type FilterField = "from" | "to" | "subject";
export type FilterOperator = "contains" | "is";
export type FilterAction = "move" | "delete";

export interface FilterRule {
  id: string;
  field: FilterField;
  operator: FilterOperator;
  value: string;
  action: FilterAction;
  targetFolder?: string;
}

export interface VacationSettings {
  enabled: boolean;
  subject: string;
  body: string;
  days: number;
}

const FILTER_BLOCK_START = "# oxmail:filter:start";
const FILTER_BLOCK_END = "# oxmail:filter:end";
const VACATION_BLOCK_START = "# oxmail:vacation:start";
const VACATION_BLOCK_END = "# oxmail:vacation:end";

function sieveQuote(value: string): string {
  return `"${value.replace(/\\/g, "\\\\").replace(/"/g, '\\"')}"`;
}

function ruleToSieve(rule: FilterRule): string {
  const test = `header :${rule.operator === "is" ? "is" : "contains"} ${sieveQuote(rule.field)} ${sieveQuote(rule.value)}`;
  const action =
    rule.action === "delete"
      ? "discard;"
      : `fileinto ${sieveQuote(rule.targetFolder ?? "INBOX")};`;
  return `if ${test} {\n    ${action}\n}`;
}

export function generateFilterScript(rules: FilterRule[]): string {
  if (rules.length === 0) return "";
  const requires = new Set(["fileinto"]);
  const body = rules.map(ruleToSieve).join("\n");
  return `require [${Array.from(requires).map(sieveQuote).join(", ")}];\n\n${FILTER_BLOCK_START}\n${body}\n${FILTER_BLOCK_END}`;
}

export function generateVacationScript(settings: VacationSettings): string {
  if (!settings.enabled) return "";
  return `require ["vacation"];\n\n${VACATION_BLOCK_START}\nif true {\n    vacation :days ${settings.days} :subject ${sieveQuote(settings.subject)} ${sieveQuote(settings.body)};\n}\n${VACATION_BLOCK_END}`;
}

function extractBlock(script: string, startMarker: string, endMarker: string): string | null {
  const startIdx = script.indexOf(startMarker);
  const endIdx = script.indexOf(endMarker);
  if (startIdx === -1 || endIdx === -1 || endIdx < startIdx) return null;
  return script.slice(startIdx + startMarker.length, endIdx).trim();
}

export function parseFilterRules(script: string): FilterRule[] {
  const block = extractBlock(script, FILTER_BLOCK_START, FILTER_BLOCK_END);
  if (!block) return [];

  const rules: FilterRule[] = [];
  const ruleRegex =
    /if\s+header\s+:(contains|is)\s+"((?:[^"\\]|\\.)*)"\s+"((?:[^"\\]|\\.)*)"\s*\{\s*(fileinto\s+"((?:[^"\\]|\\.)*)"|discard)\s*;?\s*\}/g;

  let match: RegExpExecArray | null;
  let index = 0;
  while ((match = ruleRegex.exec(block)) !== null) {
    const [, operator, field, value, actionStr, targetFolder] = match;
    const unescape = (s: string) => s.replace(/\\"/g, '"').replace(/\\\\/g, "\\");
    rules.push({
      id: `rule-${index++}`,
      field: unescape(field) as FilterField,
      operator: operator as FilterOperator,
      value: unescape(value),
      action: actionStr.startsWith("discard") ? "delete" : "move",
      targetFolder: actionStr.startsWith("discard") ? undefined : unescape(targetFolder),
    });
  }

  return rules;
}

export function parseVacationSettings(script: string): VacationSettings {
  const block = extractBlock(script, VACATION_BLOCK_START, VACATION_BLOCK_END);
  if (!block) {
    return { enabled: false, subject: "", body: "", days: 7 };
  }

  const vacationRegex =
    /vacation\s+:days\s+(\d+)\s+:subject\s+"((?:[^"\\]|\\.)*)"\s+"((?:[^"\\]|\\.)*)"/;
  const match = vacationRegex.exec(block);
  if (!match) {
    return { enabled: false, subject: "", body: "", days: 7 };
  }

  const unescape = (s: string) => s.replace(/\\"/g, '"').replace(/\\\\/g, "\\");
  const [, days, subject, body] = match;
  return {
    enabled: true,
    subject: unescape(subject),
    body: unescape(body),
    days: Number(days),
  };
}

/**
 * Merges the filter and vacation sub-scripts into a single sieve script,
 * since a mailbox only has one active script but oxmail manages both
 * concerns from separate UI pages.
 */
export function mergeSieveScripts(filterScript: string, vacationScript: string): string {
  const parts = [filterScript, vacationScript].filter((part) => part.trim().length > 0);
  if (parts.length === 0) return "";

  const requiresSet = new Set<string>();
  const bodies: string[] = [];

  for (const part of parts) {
    const requireMatch = /require\s+\[([^\]]*)\];/.exec(part);
    if (requireMatch) {
      for (const raw of requireMatch[1].split(",")) {
        const trimmed = raw.trim().replace(/^"|"$/g, "");
        if (trimmed) requiresSet.add(trimmed);
      }
    }
    const withoutRequire = part.replace(/require\s+\[[^\]]*\];\s*/, "").trim();
    if (withoutRequire) bodies.push(withoutRequire);
  }

  const requireLine = `require [${Array.from(requiresSet).map(sieveQuote).join(", ")}];`;
  return `${requireLine}\n\n${bodies.join("\n\n")}`;
}

export function splitSieveScript(script: string): { filterScript: string; vacationScript: string } {
  const filterBlock = extractBlock(script, FILTER_BLOCK_START, FILTER_BLOCK_END);
  const vacationBlock = extractBlock(script, VACATION_BLOCK_START, VACATION_BLOCK_END);

  return {
    filterScript: filterBlock
      ? `require ["fileinto"];\n\n${FILTER_BLOCK_START}\n${filterBlock}\n${FILTER_BLOCK_END}`
      : "",
    vacationScript: vacationBlock
      ? `require ["vacation"];\n\n${VACATION_BLOCK_START}\n${vacationBlock}\n${VACATION_BLOCK_END}`
      : "",
  };
}
