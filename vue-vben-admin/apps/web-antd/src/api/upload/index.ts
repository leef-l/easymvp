import { requestClient } from '#/api/request';

interface UploadResult {
  url: string;
}

/**
 * 上传文件
 */
export async function uploadFile(file: File): Promise<UploadResult> {
  const formData = new FormData();
  formData.append('file', file);
  return requestClient.post('/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
}
