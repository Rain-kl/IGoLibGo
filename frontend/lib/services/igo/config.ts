// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

import { BaseService } from '@/lib/services/core';
import type {
  BackupExportRequest,
  BackupExportResponse,
  BackupImportRequest,
  ProtocolTemplatesResponse,
  SaveProtocolTemplatesRequest,
  SaveSettingsRequest,
  SaveWebDAVRequest,
  SettingsResponse,
  WebDAVSettings,
  WebDAVSyncResponse,
} from './types';

export class IGoConfigService extends BaseService {
  protected static readonly basePath = '/api/v1/igo';

  static async getProtocolTemplates(): Promise<ProtocolTemplatesResponse> {
    return this.get<ProtocolTemplatesResponse>('/protocol/templates');
  }

  static async getDefaultProtocolTemplates(): Promise<ProtocolTemplatesResponse> {
    return this.get<ProtocolTemplatesResponse>('/protocol/templates/defaults');
  }

  static async saveProtocolTemplates(
    data: SaveProtocolTemplatesRequest,
  ): Promise<ProtocolTemplatesResponse> {
    return this.put<ProtocolTemplatesResponse>(
      '/protocol/templates',
      data as unknown as Record<string, unknown>,
    );
  }

  static async resetProtocolTemplates(): Promise<ProtocolTemplatesResponse> {
    return this.post<ProtocolTemplatesResponse>('/protocol/templates/reset');
  }

  static async getSettings(): Promise<SettingsResponse> {
    return this.get<SettingsResponse>('/settings');
  }

  static async saveSettings(
    data: SaveSettingsRequest,
  ): Promise<SettingsResponse> {
    return this.put<SettingsResponse>(
      '/settings',
      data as unknown as Record<string, unknown>,
    );
  }

  static async exportBackup(
    data: BackupExportRequest,
  ): Promise<BackupExportResponse> {
    return this.post<BackupExportResponse>(
      '/backup/export',
      data as unknown as Record<string, unknown>,
    );
  }

  static async importBackup(data: BackupImportRequest): Promise<void> {
    return this.post<void>(
      '/backup/import',
      data as unknown as Record<string, unknown>,
    );
  }

  static async getWebDAV(): Promise<WebDAVSettings> {
    return this.get<WebDAVSettings>('/webdav');
  }

  static async saveWebDAV(data: SaveWebDAVRequest): Promise<WebDAVSettings> {
    return this.put<WebDAVSettings>(
      '/webdav',
      data as unknown as Record<string, unknown>,
    );
  }

  static async syncWebDAV(): Promise<WebDAVSyncResponse> {
    return this.post<WebDAVSyncResponse>('/webdav/sync');
  }
}
