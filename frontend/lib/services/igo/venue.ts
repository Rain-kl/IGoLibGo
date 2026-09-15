// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

import { BaseService } from '@/lib/services/core';
import type {
  BoundLibraryResponse,
  DeleteSeatLabelsRequest,
  LibraryLayoutResponse,
  LibraryRuleResponse,
  LibrarySummary,
  SaveFavoritesRequest,
  SeatLabel,
  SeatRef,
  SetSeatLabelsRequest,
} from './types';

export class IGoVenueService extends BaseService {
  protected static readonly basePath = '/api/v1/igo/libraries';

  static async listLibraries(): Promise<LibrarySummary[]> {
    return this.get<LibrarySummary[]>('');
  }

  static async getBoundLibrary(): Promise<BoundLibraryResponse> {
    return this.get<BoundLibraryResponse>('/bound');
  }

  static async refreshBoundLibrary(): Promise<BoundLibraryResponse> {
    return this.post<BoundLibraryResponse>('/bound/refresh');
  }

  static async getLibrary(id: number): Promise<LibrarySummary> {
    return this.get<LibrarySummary>(`/${id}`);
  }

  static async getLibraryLayout(id: number): Promise<LibraryLayoutResponse> {
    return this.get<LibraryLayoutResponse>(`/${id}/layout`);
  }

  static async getLibraryRule(id: number): Promise<LibraryRuleResponse> {
    return this.get<LibraryRuleResponse>(`/${id}/rule`);
  }

  static async bindLibrary(id: number): Promise<BoundLibraryResponse> {
    return this.post<BoundLibraryResponse>(`/${id}/bind`);
  }

  static async previewLibrary(id: number): Promise<LibraryLayoutResponse> {
    return this.post<LibraryLayoutResponse>(`/${id}/preview`);
  }

  static async getFavorites(id: number): Promise<SeatRef[]> {
    return this.get<SeatRef[]>(`/${id}/favorites`);
  }

  static async saveFavorites(
    id: number,
    data: SaveFavoritesRequest,
  ): Promise<void> {
    return this.put<void>(
      `/${id}/favorites`,
      data as unknown as Record<string, unknown>,
    );
  }

  static async getSeatLabels(id: number): Promise<SeatLabel[]> {
    return this.get<SeatLabel[]>(`/${id}/seat-labels`);
  }

  static async setSeatLabels(
    id: number,
    data: SetSeatLabelsRequest,
  ): Promise<SeatLabel[]> {
    return this.put<SeatLabel[]>(
      `/${id}/seat-labels`,
      data as unknown as Record<string, unknown>,
    );
  }

  static async deleteSeatLabels(
    id: number,
    data: DeleteSeatLabelsRequest,
  ): Promise<void> {
    return this.delete<void>(`/${id}/seat-labels`, {
      data: data as unknown as Record<string, unknown>,
    });
  }
}
