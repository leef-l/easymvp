export interface TaskItem {
  id: string;
  title: string;
  engineCode: string;
  status: string;
  repoPath: string;
  worktreePath?: string;
  branchName?: string;
  instruction?: string;
  responseSummary?: string;
  errorMessage?: string;
  createdAt?: string;
  startedAt?: string;
  finishedAt?: string;
}

export interface TaskListParams {
  pageNum: number;
  pageSize: number;
  engineCode?: string;
  status?: string;
}

export interface TaskExecuteParams {
  title: string;
  engineCode: string;
  projectID?: string;
  repoPath: string;
  worktreePath?: string;
  branchName?: string;
  instruction: string;
}

export interface TaskLogItem {
  id: string;
  taskID: string;
  seq: number;
  logType: string;
  content: string;
  createdAt?: string;
}
