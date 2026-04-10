import { getRoleDefinitions, type RoleDefinitionItem } from '#/api/mvp/workflow';

export interface RoleTypeMeta {
  color: string;
  label: string;
}

export const fallbackRoleDefinitions: RoleDefinitionItem[] = [
  { roleType: 'architect', displayName: '架构师', color: 'purple', description: '', preferredLevels: ['max', 'pro', 'lite'], defaultSystemPrompt: '', acceptanceJudge: false, sort: 10 },
  { roleType: 'implementer', displayName: '实现者', color: 'blue', description: '', preferredLevels: ['pro', 'max', 'lite'], defaultSystemPrompt: '', acceptanceJudge: false, sort: 20 },
  { roleType: 'auditor', displayName: '审核者', color: 'green', description: '', preferredLevels: ['pro', 'max', 'lite'], defaultSystemPrompt: '', acceptanceJudge: false, sort: 30 },
  { roleType: 'coordinator', displayName: '协调者', color: 'orange', description: '', preferredLevels: ['lite', 'pro', 'max'], defaultSystemPrompt: '', acceptanceJudge: false, sort: 40 },
  { roleType: 'operator', displayName: '运维恢复师', color: 'cyan', description: '', preferredLevels: ['pro', 'max', 'lite'], defaultSystemPrompt: '', acceptanceJudge: false, sort: 50 },
  { roleType: 'experience_reviewer', displayName: '体验评审师', color: 'magenta', description: '', preferredLevels: ['max', 'pro', 'lite'], defaultSystemPrompt: '', acceptanceJudge: true, sort: 60 },
];

export async function loadRoleDefinitions(): Promise<RoleDefinitionItem[]> {
  try {
    const res = await getRoleDefinitions();
    if (res?.list?.length) {
      return [...res.list].toSorted((a, b) => (a.sort ?? 0) - (b.sort ?? 0));
    }
  } catch {
    // ignore and fallback
  }
  return fallbackRoleDefinitions;
}

export function toRoleTypeOptions(definitions: RoleDefinitionItem[]) {
  return definitions.map((item) => ({
    label: item.displayName || item.roleType,
    value: item.roleType,
  }));
}

export function toRoleTypeMap(definitions: RoleDefinitionItem[]): Record<string, RoleTypeMeta> {
  const map: Record<string, RoleTypeMeta> = {};
  for (const item of fallbackRoleDefinitions) {
    map[item.roleType] = {
      label: item.displayName || item.roleType,
      color: item.color || 'default',
    };
  }

  for (const item of definitions) {
    if (!item.roleType) {
      continue;
    }
    map[item.roleType] = {
      label: item.displayName || item.roleType,
      color: item.color || 'default',
    };
  }

  return map;
}

export async function loadRoleTypeMap() {
  return toRoleTypeMap(await loadRoleDefinitions());
}
