import type { TaskItem } from '#/api/mvp/task/types';

export interface TaskTreeOption {
  id: string;
  label: string;
  name: string;
  title: string;
  value: string;
  children?: TaskTreeOption[];
}

export function toTaskTreeOptions(tasks: TaskItem[]): TaskTreeOption[] {
  return tasks.map((task) => ({
    id: task.id,
    label: task.name,
    name: task.name,
    title: task.name,
    value: task.id,
    children: task.children?.length ? toTaskTreeOptions(task.children) : undefined,
  }));
}
