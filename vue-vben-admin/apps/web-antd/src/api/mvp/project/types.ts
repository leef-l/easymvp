/** MVP项目表类型定义 */

/** MVP项目表项 */
export interface ProjectItem {
  id: string;
  name: string;
  projectCategory?: string;
  description?: string;
  status?: string;
  pauseReason?: string;
  globalContext?: string;
  architectModelID?: string;
  workDir?: string;
  createdAt?: string;
  updatedAt?: string;
}

/** MVP项目表列表查询参数 */
export interface ProjectListParams {
  pageNum: number;
  pageSize: number;
  orderBy?: string;
  orderDir?: string;
  startTime?: string;
  endTime?: string;
  name?: string;
}

/** MVP项目表创建参数 */
export interface ProjectCreateParams {
  name: string;
  projectCategory: string;
  description?: string;
  status?: string;
  pauseReason?: string;
  globalContext?: string;
  architectModelID?: string;
}

/** MVP项目表更新参数 */
export interface ProjectUpdateParams {
  id: string;
  name: string;
  projectCategory?: string;
  description?: string;
  status?: string;
  pauseReason?: string;
  globalContext?: string;
  architectModelID?: string;
  workDir?: string;
}
