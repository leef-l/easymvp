import { requestClient } from '#/api/request';

import type { EngineDetailItem, EngineItem, EngineUpdateParams } from './types';

const PREFIX = '/ai/engine';

export function getEngineList() {
  return requestClient.get<{ list: EngineItem[] }>(`${PREFIX}/list`);
}

export function getEngineDetail(engineCode: string) {
  return requestClient.get<EngineDetailItem>(`${PREFIX}/detail`, {
    params: { engineCode },
  });
}

export function updateEngine(data: EngineUpdateParams) {
  return requestClient.post(`${PREFIX}/update`, data);
}

export function testEngineConnection(engineCode: string) {
  return requestClient.post<{ success: boolean; message: string }>(
    `${PREFIX}/test-connection`,
    { engineCode },
  );
}
