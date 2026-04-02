/** 角色类型选项 */
export const roleTypeOptions = [
  { label: '架构师', value: 'architect' },
  { label: '实现者', value: 'implementer' },
  { label: '审核者', value: 'auditor' },
  { label: '协调者', value: 'coordinator' },
];

/** 角色类型映射（用于列表/详情展示） */
export const roleTypeMap: Record<string, { label: string; color: string }> = {
  architect: { label: '架构师', color: 'purple' },
  implementer: { label: '实现者', color: 'blue' },
  auditor: { label: '审核者', color: 'green' },
  coordinator: { label: '协调者', color: 'orange' },
};

/** 角色等级选项 */
export const roleLevelOptions = [
  { label: 'Lite - 轻量级', value: 'lite' },
  { label: 'Pro - 专业级', value: 'pro' },
  { label: 'Max - 旗舰级', value: 'max' },
];

/** 角色等级映射（用于列表/详情展示） */
export const roleLevelMap: Record<string, { label: string; color: string }> = {
  lite: { label: 'Lite', color: 'default' },
  pro: { label: 'Pro', color: 'blue' },
  max: { label: 'Max', color: 'gold' },
};
