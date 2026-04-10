/** 任务日志表类型定义 */

/** 任务日志表项 */
export interface TaskLogItem {
  id: string;
  taskID: string;
  taskName?: string;
  action: string;
  fromStatus?: string;
  toStatus?: string;
  message?: string;
  operator?: string;
  createdAt?: string;
  updatedAt?: string;
}

/** 任务日志表列表查询参数 */
export interface TaskLogListParams {
  pageNum: number;
  pageSize: number;
  orderBy?: string;
  orderDir?: string;
  startTime?: string;
  endTime?: string;
}

/** 任务日志表创建参数 */
export interface TaskLogCreateParams {
  taskID?: string;
  action?: string;
  fromStatus?: string;
  toStatus?: string;
  message?: string;
  operator?: string;
}

/** 任务日志表更新参数 */
export interface TaskLogUpdateParams {
  id: string;
  taskID?: string;
  action?: string;
  fromStatus?: string;
  toStatus?: string;
  message?: string;
  operator?: string;
}
