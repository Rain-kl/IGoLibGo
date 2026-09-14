// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Building2 } from 'lucide-react';
import { toast } from 'sonner';

import { IGoService } from '@/lib/services/igo';
import type {
  BoundLibraryResponse,
  LibraryLayoutResponse,
  LibraryRuleResponse,
  LibrarySummary,
  SeatRef,
  SeatSnapshot,
  SessionResponse,
} from '@/lib/services/igo/types';

import { SessionAuthCard } from '@/components/igo/venue/session-auth-card';
import { VenueSelectorCard } from '@/components/igo/venue/venue-selector-card';
import { VenueRuleCard } from '@/components/igo/venue/venue-rule-card';
import { SeatLayoutGrid } from '@/components/igo/venue/seat-layout-grid';
import { FavoritesDrawer } from '@/components/igo/venue/favorites-drawer';
import { SeatLabelDialog } from '@/components/igo/venue/seat-label-dialog';

export default function VenuePage() {
  const t = useTranslations('igo.venue');
  const tCommon = useTranslations('common');

  // 会话状态
  const [session, setSession] = React.useState<SessionResponse | null>(null);
  const [sessionLoading, setSessionLoading] = React.useState(true);

  // 场馆与列表
  const [boundInfo, setBoundInfo] = React.useState<BoundLibraryResponse | null>(
    null,
  );
  const [libraries, setLibraries] = React.useState<LibrarySummary[]>([]);
  const [venueLoading, setVenueLoading] = React.useState(true);

  // 布局与规则
  const [currentLibId, setCurrentLibId] = React.useState<number | null>(null);
  const [layout, setLayout] = React.useState<LibraryLayoutResponse | null>(
    null,
  );
  const [layoutLoading, setLayoutLoading] = React.useState(false);
  const [rule, setRule] = React.useState<LibraryRuleResponse | null>(null);
  const [ruleLoading, setRuleLoading] = React.useState(false);

  // 座位选座、收藏与备注
  const [selectedSeats, setSelectedSeats] = React.useState<SeatSnapshot[]>([]);
  const [favorites, setFavorites] = React.useState<SeatRef[]>([]);
  const [labels, setLabels] = React.useState<Record<string, string>>({});
  const [favoritesOpen, setFavoritesOpen] = React.useState(false);
  const [labelsOpen, setLabelsOpen] = React.useState(false);

  // 1. 加载会话
  const loadSession = React.useCallback(async () => {
    setSessionLoading(true);
    try {
      const res = await IGoService.session.getSession();
      setSession(res);
    } catch {
      setSession(null);
    } finally {
      setSessionLoading(false);
    }
  }, []);

  // 2. 加载绑定场馆与场馆列表
  const loadVenues = React.useCallback(async () => {
    setVenueLoading(true);
    try {
      const [boundRes, libsRes] = await Promise.all([
        IGoService.venue.getBoundLibrary(),
        IGoService.venue.listLibraries(),
      ]);
      setBoundInfo(boundRes);
      setLibraries(libsRes || []);

      if (boundRes?.library?.library_id) {
        setCurrentLibId(boundRes.library.library_id);
      }
    } catch {
      // 未登录或错误
    } finally {
      setVenueLoading(false);
    }
  }, []);

  // 3. 加载选定场馆的布局、规则与收藏
  const loadLibraryDetails = React.useCallback(async (libId: number) => {
    setLayoutLoading(true);
    setRuleLoading(true);
    setSelectedSeats([]);

    try {
      const [layoutRes, ruleRes, favsRes] = await Promise.all([
        IGoService.venue.getLibraryLayout(libId),
        IGoService.venue.getLibraryRule(libId).catch(() => null),
        IGoService.venue.getFavorites(libId).catch(() => []),
      ]);
      setLayout(layoutRes);
      setRule(ruleRes);
      setFavorites(favsRes || []);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '加载场馆详情失败');
    } finally {
      setLayoutLoading(false);
      setRuleLoading(false);
    }
  }, []);

  // 初始挂载加载
  React.useEffect(() => {
    loadSession();
    loadVenues();
  }, [loadSession, loadVenues]);

  // 当锁定场馆确定时拉取布局
  React.useEffect(() => {
    if (currentLibId) {
      loadLibraryDetails(currentLibId);
    }
  }, [currentLibId, loadLibraryDetails]);

  // 保存当前选中的座位到收藏
  const handleSaveFavorites = async () => {
    if (!currentLibId || selectedSeats.length === 0) return;
    const newItems: SeatRef[] = selectedSeats.map((s) => ({
      seat_key: s.seat_key,
      seat_name: s.seat_name || s.seat_key,
    }));

    // 合并现有收藏并去重
    const merged = [...favorites];
    for (const item of newItems) {
      if (!merged.some((f) => f.seat_key === item.seat_key)) {
        merged.push(item);
      }
    }

    try {
      const updated = await IGoService.venue.saveFavorites(currentLibId, {
        seats: merged,
      });
      setFavorites(updated);
      toast.success(tCommon('save'));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    }
  };

  // 从抽屉移除单个收藏
  const handleRemoveFavorite = async (seatKey: string) => {
    if (!currentLibId) return;
    const updated = favorites.filter((f) => f.seat_key !== seatKey);
    try {
      const res = await IGoService.venue.saveFavorites(currentLibId, {
        seats: updated,
      });
      setFavorites(res);
      toast.success(tCommon('delete'));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    }
  };

  // 保存备注标签
  const handleSaveLabels = async (text: string) => {
    if (!currentLibId || selectedSeats.length === 0) return;
    const seatsPayload: SeatRef[] = selectedSeats.map((s) => ({
      seat_key: s.seat_key,
      seat_name: s.seat_name || s.seat_key,
    }));

    const res = await IGoService.venue.setSeatLabels(currentLibId, {
      seats: seatsPayload,
      text,
    });

    const nextLabels = { ...labels };
    for (const item of res) {
      nextLabels[item.seat_key] = item.text;
    }
    setLabels(nextLabels);
  };

  // 删除备注标签
  const handleDeleteLabels = async (seatKeys: string[]) => {
    if (!currentLibId || seatKeys.length === 0) return;
    await IGoService.venue.deleteSeatLabels(currentLibId, {
      seat_keys: seatKeys,
    });

    const nextLabels = { ...labels };
    for (const k of seatKeys) {
      delete nextLabels[k];
    }
    setLabels(nextLabels);
  };

  return (
    <div className='py-6 px-1 space-y-6 w-full'>
      {/* 1. 标准标题 */}
      <div className='flex items-center gap-2'>
        <Building2 className='size-5 text-primary' />
        <div>
          <h1 className='text-2xl font-semibold tracking-tight'>
            {t('title')}
          </h1>
          <p className='text-sm text-muted-foreground mt-0.5'>
            {t('description')}
          </p>
        </div>
      </div>

      {/* 2. 账号授权卡片 */}
      <SessionAuthCard
        session={session}
        loading={sessionLoading}
        onRefresh={() => {
          loadSession();
          loadVenues();
        }}
      />

      {/* 3. 锁定场馆选择与详情卡片 */}
      <VenueSelectorCard
        boundInfo={boundInfo}
        libraries={libraries}
        loading={venueLoading}
        onRefresh={loadVenues}
        onSelectLibrary={(id) => {
          setCurrentLibId(id);
        }}
      />

      {/* 4. 场馆规则卡片 */}
      {rule && <VenueRuleCard rule={rule} loading={ruleLoading} />}

      {/* 5. 座位图可视化排布 */}
      <SeatLayoutGrid
        layout={layout}
        loading={layoutLoading}
        selectedSeats={selectedSeats}
        favorites={favorites}
        labels={labels}
        onSelectionChange={setSelectedSeats}
        onOpenFavorites={() => setFavoritesOpen(true)}
        onOpenLabels={() => setLabelsOpen(true)}
        onSaveFavorites={handleSaveFavorites}
      />

      {/* 6. 收藏抽屉 */}
      <FavoritesDrawer
        open={favoritesOpen}
        onOpenChange={setFavoritesOpen}
        favorites={favorites}
        allSeats={layout?.seats || []}
        onSelectSeat={(seat) => {
          if (!selectedSeats.some((s) => s.seat_key === seat.seat_key)) {
            setSelectedSeats([...selectedSeats, seat]);
          }
        }}
        onRemoveFavorite={handleRemoveFavorite}
      />

      {/* 7. 标签设置弹窗 */}
      <SeatLabelDialog
        open={labelsOpen}
        onOpenChange={setLabelsOpen}
        selectedSeats={selectedSeats}
        currentLabels={labels}
        onSaveLabel={handleSaveLabels}
        onDeleteLabel={handleDeleteLabels}
      />
    </div>
  );
}
