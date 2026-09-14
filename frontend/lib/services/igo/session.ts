// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

import { BaseService } from '@/lib/services/core';
import type {
  AuthFromCodeRequest,
  AuthFromCookieRequest,
  QRCodeResponse,
  SessionResponse,
  SessionWorkflowResponse,
} from './types';

export class IGoSessionService extends BaseService {
  protected static readonly basePath = '/api/v1/igo/session';

  static async getSession(): Promise<SessionResponse> {
    return this.get<SessionResponse>('');
  }

  static async getAuthQRCode(): Promise<QRCodeResponse> {
    return this.get<QRCodeResponse>('/auth-qrcode');
  }

  static async authenticateFromCode(
    data: AuthFromCodeRequest,
  ): Promise<SessionWorkflowResponse> {
    return this.post<SessionWorkflowResponse>(
      '/from-code',
      data as unknown as Record<string, unknown>,
    );
  }

  static async authenticateFromCookie(
    data: AuthFromCookieRequest,
  ): Promise<SessionWorkflowResponse> {
    return this.post<SessionWorkflowResponse>(
      '/from-cookie',
      data as unknown as Record<string, unknown>,
    );
  }

  static async refreshCookie(): Promise<SessionWorkflowResponse> {
    return this.post<SessionWorkflowResponse>('/cookie/refresh');
  }

  static async signOut(): Promise<void> {
    return this.delete<void>('');
  }
}
