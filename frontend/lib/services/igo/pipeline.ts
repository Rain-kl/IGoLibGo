// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

import { BaseService } from '@/lib/services/core';
import type {
  CreatePipelineConfigRequest,
  LibraryLayoutResponse,
  PipelineConfigDTO,
  PipelineExecutionResult,
  PipelineHelperVerifyCheckinRequest,
  PipelineHelperVerifyCheckinResponse,
  PipelineHelperVerifySessionRequest,
  PipelineHelperVerifySessionResponse,
  RunPipelineRequest,
  UpdatePipelineConfigRequest,
} from './types';

export class IGoPipelineService extends BaseService {
  protected static readonly basePath = '/api/v1/igo/pipeline';

  /**
   * List all pipeline configs for the authenticated user.
   */
  static async listConfigs(): Promise<PipelineConfigDTO[]> {
    return this.get<PipelineConfigDTO[]>('/configs');
  }

  /**
   * Create a new pipeline config card.
   */
  static async createConfig(
    data: CreatePipelineConfigRequest,
  ): Promise<PipelineConfigDTO> {
    return this.post<PipelineConfigDTO>(
      '/configs',
      data as unknown as Record<string, unknown>,
    );
  }

  /**
   * Get a single pipeline config by ID.
   */
  static async getConfig(id: string): Promise<PipelineConfigDTO> {
    return this.get<PipelineConfigDTO>(`/configs/${encodeURIComponent(id)}`);
  }

  /**
   * Update an existing pipeline config.
   */
  static async updateConfig(
    id: string,
    data: UpdatePipelineConfigRequest,
  ): Promise<PipelineConfigDTO> {
    return this.put<PipelineConfigDTO>(
      `/configs/${encodeURIComponent(id)}`,
      data as unknown as Record<string, unknown>,
    );
  }

  /**
   * Delete a pipeline config by ID.
   */
  static async deleteConfig(id: string): Promise<void> {
    return this.delete<void>(`/configs/${encodeURIComponent(id)}`);
  }

  /**
   * Execute an all-in-one pipeline immediately.
   */
  static async runConfig(
    id: string,
    data?: RunPipelineRequest,
  ): Promise<PipelineExecutionResult> {
    return this.post<PipelineExecutionResult>(
      `/configs/${encodeURIComponent(id)}/run`,
      (data || {}) as unknown as Record<string, unknown>,
    );
  }

  /**
   * Helper: verify or exchange session credentials.
   */
  static async helperVerifySession(
    data: PipelineHelperVerifySessionRequest,
  ): Promise<PipelineHelperVerifySessionResponse> {
    return this.post<PipelineHelperVerifySessionResponse>(
      '/helpers/verify-session',
      data as unknown as Record<string, unknown>,
    );
  }

  /**
   * Helper: fetch library layout using specific credentials.
   */
  static async helperGetLibraryLayout(params?: {
    cookie?: string;
    auth_code?: string;
    auth_url?: string;
  }): Promise<LibraryLayoutResponse[]> {
    return this.get<LibraryLayoutResponse[]>('/helpers/library-layout', {
      params,
    });
  }

  /**
   * Helper: verify or exchange checkin token.
   */
  static async helperVerifyCheckin(
    data: PipelineHelperVerifyCheckinRequest,
  ): Promise<PipelineHelperVerifyCheckinResponse> {
    return this.post<PipelineHelperVerifyCheckinResponse>(
      '/helpers/verify-checkin',
      data as unknown as Record<string, unknown>,
    );
  }
}
