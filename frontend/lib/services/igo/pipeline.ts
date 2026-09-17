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
      { timeout: 180000 },
    );
  }

  /**
   * Helper: verify or exchange session credentials.
   */
  static async helperVerifySession(
    data: PipelineHelperVerifySessionRequest,
  ): Promise<PipelineHelperVerifySessionResponse> {
    const input =
      data.cookie?.trim() ||
      data.auth_url?.trim() ||
      data.auth_code?.trim() ||
      '';
    return this.post<PipelineHelperVerifySessionResponse>('/verify-session', {
      cookie: input,
      account_id: data.account_id,
    });
  }

  /**
   * Helper: fetch library layout using specific credentials.
   */
  static async helperGetLibraryLayout(data: {
    cookie?: string;
    account_id?: string;
    library_id: number;
    auth_code?: string;
    auth_url?: string;
  }): Promise<LibraryLayoutResponse> {
    const cookieInput =
      data.cookie?.trim() ||
      data.auth_url?.trim() ||
      data.auth_code?.trim() ||
      '';
    return this.post<LibraryLayoutResponse>('/library-layout', {
      cookie: cookieInput,
      account_id: data.account_id,
      library_id: data.library_id,
    });
  }

  /**
   * Helper: verify or exchange checkin token.
   */
  static async helperVerifyCheckin(
    data: PipelineHelperVerifyCheckinRequest,
  ): Promise<PipelineHelperVerifyCheckinResponse> {
    const input =
      data.checkin_token?.trim() ||
      data.checkin_url?.trim() ||
      data.checkin_code?.trim() ||
      '';
    return this.post<PipelineHelperVerifyCheckinResponse>('/verify-checkin', {
      token_or_code: input,
    });
  }
}
