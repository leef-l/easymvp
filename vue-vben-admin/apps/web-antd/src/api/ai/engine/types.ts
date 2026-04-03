export interface EngineItem {
  id: string;
  code: string;
  name: string;
  description?: string;
  status?: number;
  configStatus?: number;
  defaultModelID?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface EngineDetailItem extends EngineItem {
  engineCode: string;
  baseURL?: string;
  apiKeyMasked?: string;
  timeoutSeconds?: number;
  maxSteps?: number;
  workspaceRoot?: string;
  commandTemplate?: string;
  callbackURL?: string;
  callbackSecret?: string;
  extraConfig?: string;
}

export interface EngineUpdateParams {
  engineCode: string;
  defaultModelID?: string;
  timeoutSeconds?: number;
  maxSteps?: number;
  workspaceRoot?: string;
  commandTemplate?: string;
  callbackURL?: string;
  callbackSecret?: string;
  extraConfig?: Record<string, any>;
  status?: number;
}
