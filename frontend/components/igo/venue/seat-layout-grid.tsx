// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import {
  Bookmark,
  BookmarkCheck,
  CheckSquare,
  Grid3X3,
  Rocket,
  Tag,
  Trash2,
  ZoomIn,
  ZoomOut,
} from 'lucide-react';
import { toast } from 'sonner';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Spinner } from '@/components/ui/spinner';
import { SeatNode } from './seat-node';
import type {
  LibraryLayoutResponse,
  SeatLabel,
  SeatRef,
  SeatSnapshot,
} from '@/lib/services/igo/types';

interface SeatLayoutGridProps {
  layout: LibraryLayoutResponse | null;
  loading: boolean;
  selectedSeats: SeatSnapshot[];
  favorites: SeatRef[];
  labels: Record<string, string>;
  onSelectionChange: (seats: SeatSnapshot[]) => void;
  onOpenFavorites: () => void;
  onOpenLabels: () => void;
  onSaveFavorites: () => void;
}

export function SeatLayoutGrid({
  layout,
  loading,
  selectedSeats,
  favorites,
  labels,
  onSelectionChange,
  onOpenFavorites,
  onOpenLabels,
  onSaveFavorites,
}: SeatLayoutGridProps) {
  const t = useTranslations('igo.venue');
  const tCommon = useTranslations('common');
  const router = useRouter();

  const [zoomLevel, setZoomLevel] = React.useState<number>(1);

  // 映射收藏 Set
  const favoriteKeysSet = React.useMemo(() => {
    return new Set(favorites.map((f) => f.seat_key));
  }, [favorites]);

  // 计算网格边界
  const { seatsByCoord, maxX, maxY } = React.useMemo(() => {
    if (!layout || !layout.seats || layout.seats.length === 0) {
      return {
        seatsByCoord: new Map<string, SeatSnapshot>(),
        maxX: 0,
        maxY: 0,
      };
    }

    const map = new Map<string, SeatSnapshot>();
    let mx = layout.max_x ?? 0;
    let my = layout.max_y ?? 0;

    for (const seat of layout.seats) {
      map.set(`${seat.x},${seat.y}`, seat);
      if (seat.x > mx) mx = seat.x;
      if (seat.y > my) my = seat.y;
    }

    return { seatsByCoord: map, maxX: mx, maxY: my };
  }, [layout]);

  // 切换座位选中状态
  const handleToggleSeat = (seat: SeatSnapshot) => {
    const existsIndex = selectedSeats.findIndex(
      (s) => s.seat_key === seat.seat_key,
    );
    if (existsIndex >= 0) {
      onSelectionChange(
        selectedSeats.filter((s) => s.seat_key !== seat.seat_key),
      );
    } else {
      onSelectionChange([...selectedSeats, seat]);
    }
  };

  // 清空选中
  const handleClearSelection = () => {
    onSelectionChange([]);
  };

  // 设为抢座目标并跳转
  const handleApplyGrab = () => {
    if (selectedSeats.length === 0) return;
    try {
      // 存入 sessionStorage 供抢座页快速读取
      const grabCandidates = selectedSeats.map((s) => ({
        seat_key: s.seat_key,
        seat_name: s.seat_name || s.seat_key,
      }));
      sessionStorage.setItem(
        'igo_grab_candidates',
        JSON.stringify(grabCandidates),
      );
      router.push('/grab');
    } catch {
      router.push('/grab');
    }
  };

  if (loading) {
    return (
      <Card className='border-border/60 shadow-sm'>
        <CardContent className='py-24 flex flex-col items-center justify-center gap-3'>
          <Spinner className='size-8' />
          <p className='text-xs text-muted-foreground'>{tCommon('loading')}</p>
        </CardContent>
      </Card>
    );
  }

  if (!layout || !layout.seats || layout.seats.length === 0) {
    return (
      <Card className='border-dashed shadow-none'>
        <CardContent className='py-16 text-center text-xs text-muted-foreground'>
          请先在上方锁定场馆以加载座位可视化排布图
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className='border-border/60 shadow-sm'>
      <CardHeader className='pb-3'>
        <div className='flex flex-wrap items-center justify-between gap-2'>
          <div className='flex items-center gap-2'>
            <Grid3X3 className='size-4 text-primary' />
            <CardTitle className='text-base font-semibold'>
              {t('seatGridTitle')}
            </CardTitle>
            <Badge variant='outline' className='text-xs font-normal'>
              {layout.available_seats} 可选 / 共 {layout.total_seats} 座
            </Badge>
          </div>

          {/* 快捷操作与收藏抽屉入口 */}
          <div className='flex items-center gap-1.5'>
            <Button
              variant='outline'
              size='sm'
              onClick={onOpenFavorites}
              className='gap-1 h-8 text-xs'
            >
              <Bookmark className='size-3 text-amber-500' />
              <span>
                {t('favoritesBtn')} ({favorites.length})
              </span>
            </Button>
            <Button
              variant='outline'
              size='sm'
              onClick={onOpenLabels}
              className='gap-1 h-8 text-xs'
            >
              <Tag className='size-3 text-primary' />
              <span>{t('labelsBtn')}</span>
            </Button>
            <div className='flex items-center border rounded-md p-0.5 ml-2'>
              <Button
                variant='ghost'
                size='icon'
                aria-label='缩小座位图'
                onClick={() => setZoomLevel((z) => Math.max(0.75, z - 0.1))}
                className='size-7'
              >
                <ZoomOut className='size-3.5' />
              </Button>
              <span className='text-[10px] font-mono px-1 select-none text-muted-foreground'>
                {Math.round(zoomLevel * 100)}%
              </span>
              <Button
                variant='ghost'
                size='icon'
                aria-label='放大座位图'
                onClick={() => setZoomLevel((z) => Math.min(1.5, z + 0.1))}
                className='size-7'
              >
                <ZoomIn className='size-3.5' />
              </Button>
            </div>
          </div>
        </div>

        {/* 图例栏 */}
        <div className='flex flex-wrap items-center gap-4 text-[11px] text-muted-foreground pt-2'>
          <div className='flex items-center gap-1.5'>
            <span className='size-3.5 rounded border border-border bg-card' />
            <span>{t('legendAvailable')}</span>
          </div>
          <div className='flex items-center gap-1.5'>
            <span className='size-3.5 rounded bg-muted/60 border border-border/40' />
            <span>{t('legendOccupied')}</span>
          </div>
          <div className='flex items-center gap-1.5'>
            <span className='size-3.5 rounded bg-primary text-primary-foreground flex items-center justify-center font-bold text-[8px]'>
              1
            </span>
            <span>{t('legendSelected')}</span>
          </div>
          <div className='flex items-center gap-1.5'>
            <Bookmark className='size-3 fill-amber-500 text-amber-500' />
            <span>已收藏</span>
          </div>
          <div className='flex items-center gap-1.5'>
            <Tag className='size-3 text-primary' />
            <span>有备注</span>
          </div>
        </div>
      </CardHeader>

      <CardContent className='space-y-4'>
        {/* 选座工具栏 */}
        {selectedSeats.length > 0 && (
          <div className='flex flex-wrap items-center justify-between gap-2 p-2.5 rounded-lg bg-primary/5 border border-primary/20 text-xs'>
            <div className='flex items-center gap-2'>
              <CheckSquare className='size-4 text-primary' />
              <span className='font-medium text-foreground'>
                {t('selectedCount', { count: selectedSeats.length })}:
              </span>
              <div className='flex flex-wrap gap-1 max-w-lg'>
                {selectedSeats.slice(0, 8).map((s, idx) => (
                  <Badge
                    key={s.seat_key}
                    variant='secondary'
                    className='text-[10px] font-mono py-0 h-5'
                  >
                    #{idx + 1} {s.seat_name || s.seat_key}
                  </Badge>
                ))}
                {selectedSeats.length > 8 && (
                  <span className='text-[10px] text-muted-foreground self-center'>
                    +{selectedSeats.length - 8}
                  </span>
                )}
              </div>
            </div>

            <div className='flex items-center gap-2'>
              <Button
                variant='ghost'
                size='sm'
                onClick={handleClearSelection}
                className='h-7 text-xs text-muted-foreground'
              >
                <Trash2 className='size-3 mr-1' />
                {t('clearSelection')}
              </Button>
              <Button
                variant='outline'
                size='sm'
                onClick={onSaveFavorites}
                className='h-7 text-xs gap-1'
              >
                <BookmarkCheck className='size-3 text-amber-500' />
                {t('saveToFavorites')}
              </Button>
              <Button
                variant='default'
                size='sm'
                onClick={handleApplyGrab}
                className='h-7 text-xs gap-1'
              >
                <Rocket className='size-3' />
                {t('applyGrab')}
              </Button>
            </div>
          </div>
        )}

        {/* 座位排布网格渲染 */}
        <div className='border rounded-lg bg-muted/15 p-4 overflow-auto min-h-96 max-h-[640px]'>
          <div
            style={{
              transform: `scale(${zoomLevel})`,
              transformOrigin: 'top left',
              transition: 'transform 0.15s ease-out',
            }}
            className='inline-block'
          >
            <div
              className='grid gap-1.5'
              style={{
                gridTemplateColumns: `repeat(${Math.max(maxX, 1)}, minmax(40px, 1fr))`,
              }}
            >
              {Array.from({ length: maxY }, (_, yIdx) => {
                const y = yIdx + 1;
                return Array.from({ length: maxX }, (_, xIdx) => {
                  const x = xIdx + 1;
                  const seat = seatsByCoord.get(`${x},${y}`);

                  if (!seat) {
                    // 空过道或空白占位
                    return (
                      <div key={`empty-${x}-${y}`} className='w-10 h-10' />
                    );
                  }

                  const selectedIdx = selectedSeats.findIndex(
                    (s) => s.seat_key === seat.seat_key,
                  );

                  return (
                    <SeatNode
                      key={seat.seat_key}
                      seat={seat}
                      isSelected={selectedIdx >= 0}
                      selectionIndex={
                        selectedIdx >= 0 ? selectedIdx : undefined
                      }
                      label={labels[seat.seat_key]}
                      isFavorite={favoriteKeysSet.has(seat.seat_key)}
                      onToggle={handleToggleSeat}
                    />
                  );
                });
              })}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
