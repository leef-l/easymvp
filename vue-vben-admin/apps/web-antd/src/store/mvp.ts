import { ref, computed } from 'vue';
import { defineStore } from 'pinia';

/**
 * MVP 项目状态全局 Store
 * 统一管理项目状态、任务统计、SSE 事件通知
 * 替代各页面组件中的本地轮询
 */
export const useMvpStore = defineStore('mvp', () => {
  // ==================== 当前选中的项目 ====================

  const currentProjectId = ref<string>('');
  const currentProjectName = ref<string>('');

  // ==================== 项目状态数据 ====================

  interface ProjectStatusData {
    status: string;
    pauseReason?: string;
    activeBatch: number;
    totalTasks: number;
    statusCounts: Record<string, number>;
    lastActiveAt?: string;
    isActuallyWorking: boolean;
    activeRunningTasks: number;
    stalledTaskCount: number;
  }

  const projectStatus = ref<ProjectStatusData | null>(null);
  const isLoadingStatus = ref(false);
  const lastStatusUpdate = ref<number>(0);

  // ==================== 计算属性 ====================

  const completedTasks = computed(() => projectStatus.value?.statusCounts?.completed ?? 0);
  const totalTasks = computed(() => projectStatus.value?.totalTasks ?? 0);
  const progressPercent = computed(() => {
    if (totalTasks.value === 0) return 0;
    return Math.floor((completedTasks.value / totalTasks.value) * 100);
  });
  const isRunning = computed(() => projectStatus.value?.status === 'running');
  const isPaused = computed(() => projectStatus.value?.status === 'paused');
  const isCompleted = computed(() => projectStatus.value?.status === 'completed');

  // ==================== 状态更新 ====================

  /** 更新项目状态（从 API 或 SSE 推送） */
  function updateProjectStatus(data: ProjectStatusData) {
    projectStatus.value = data;
    lastStatusUpdate.value = Date.now();
  }

  /** 切换当前项目 */
  function setCurrentProject(id: string, name: string) {
    currentProjectId.value = id;
    currentProjectName.value = name;
    projectStatus.value = null;
  }

  /** 清除项目状态 */
  function clearProjectStatus() {
    currentProjectId.value = '';
    currentProjectName.value = '';
    projectStatus.value = null;
  }

  // ==================== 事件总线（组件间通信） ====================

  type EventHandler = (...args: any[]) => void;
  const eventHandlers = new Map<string, Set<EventHandler>>();

  /** 订阅事件 */
  function on(event: string, handler: EventHandler) {
    if (!eventHandlers.has(event)) {
      eventHandlers.set(event, new Set());
    }
    eventHandlers.get(event)!.add(handler);

    // 返回取消订阅函数
    return () => {
      eventHandlers.get(event)?.delete(handler);
    };
  }

  /** 触发事件 */
  function emit(event: string, ...args: any[]) {
    eventHandlers.get(event)?.forEach((handler) => {
      try {
        handler(...args);
      } catch (e) {
        console.error(`[MvpStore] 事件处理异常: ${event}`, e);
      }
    });
  }

  // 预定义事件名称
  const Events = {
    TASK_COMPLETED: 'task:completed',
    TASK_FAILED: 'task:failed',
    BATCH_ADVANCED: 'batch:advanced',
    PROJECT_COMPLETED: 'project:completed',
    STATUS_UPDATED: 'status:updated',
  } as const;

  return {
    // State
    currentProjectId,
    currentProjectName,
    projectStatus,
    isLoadingStatus,
    lastStatusUpdate,

    // Computed
    completedTasks,
    totalTasks,
    progressPercent,
    isRunning,
    isPaused,
    isCompleted,

    // Actions
    updateProjectStatus,
    setCurrentProject,
    clearProjectStatus,

    // Events
    on,
    emit,
    Events,
  };
});
