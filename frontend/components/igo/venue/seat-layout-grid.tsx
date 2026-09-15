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
  Move,
  Rocket,
  RotateCcw,
  Tag,
  Trash2,
  ZoomIn,
  ZoomOut,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Spinner } from '@/components/ui/spinner';
import { SeatNode } from './seat-node';
import type {
  LibraryLayoutResponse,
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
  selectedSeats = [],
  favorites = [],
  labels = {},
  onSelectionChange,
  onOpenFavorites,
  onOpenLabels,
  onSaveFavorites,
}: SeatLayoutGridProps) {
  const t = useTranslations('igo.venue');
  const tCommon = useTranslations('common');
  const router = useRouter();

  const [zoomLevel, setZoomLevel] = React.useState<number>(1);
  const [panPosition, setPanPosition] = React.useState<{
    x: number;
    y: number;
  }>({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = React.useState(false);
  const dragStartRef = React.useRef<{ x: number; y: number }>({ x: 0, y: 0 });
  const panStartRef = React.useRef<{ x: number; y: number }>({ x: 0, y: 0 });

  // 映射收藏 Set
  const favoriteKeysSet = React.useMemo(() => {
    return new Set((favorites || []).map((f) => f.seat_key));
  }, [favorites]);

  // 计算紧凑映射网格边界：将稀疏坐标映射到紧凑连续的行/列索引
  const { seatsByCoord, totalCols, totalRows } = React.useMemo(() => {
    if (!layout || !layout.seats || layout.seats.length === 0) {
      return {
        seatsByCoord: new Map<string, SeatSnapshot>(),
        totalCols: 0,
        totalRows: 0,
      };
    }

    // 提取所有唯一且排序的 X 和 Y
    const xs = Array.from(new Set(layout.seats.map((s) => s.x))).sort(
      (a, b) => a - b,
    );
    const ys = Array.from(new Set(layout.seats.map((s) => s.y))).sort(
      (a, b) => a - b,
    );

    const xMap = new Map<number, number>();
    xs.forEach((x, idx) => xMap.set(x, idx + 1));

    const yMap = new Map<number, number>();
    ys.forEach((y, idx) => yMap.set(y, idx + 1));

    const map = new Map<string, SeatSnapshot>();
    for (const seat of layout.seats) {
      const col = xMap.get(seat.x) ?? 1;
      const row = yMap.get(seat.y) ?? 1;
      map.set(`${col},${row}`, seat);
    }

    return {
      seatsByCoord: map,
      totalCols: xs.length,
      totalRows: ys.length,
    };
  }, [layout]);

  // 鼠标拖动画布事件处理
  const handleMouseDown = (e: React.MouseEvent<HTMLDivElement>) => {
    // 仅响应鼠标左键或中键按在画布背景上
    if (e.button !== 0 && e.button !== 1) return;
    // 如果点击目标是按钮或者座位，不启动拖拽
    if ((e.target as HTMLElement).closest('button')) return;

    setIsDragging(true);
    dragStartRef.current = { x: e.clientX, y: e.clientY };
    panStartRef.current = { ...panPosition };
  };

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!isDragging) return;
    const dx = e.clientX - dragStartRef.current.x;
    const dy = e.clientY - dragStartRef.current.y;
    setPanPosition({
      x: panStartRef.current.x + dx,
      y: panStartRef.current.y + dy,
    });
  };

  const handleMouseUpOrLeave = () => {
    if (isDragging) {
      setIsDragging(false);
    }
  };

  const handleResetView = () => {
    setZoomLevel(1);
    setPanPosition({ x: 0, y: 0 });
  };

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
      <Card className='border-dashed shadow-none'>
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
    <Card className='border-dashed shadow-none'>
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
                {t('favoritesBtn')} ({favorites?.length || 0})
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
            <div className='flex items-center border rounded-md p-0.5 ml-2 bg-background'>
              <Button
                variant='ghost'
                size='icon'
                aria-label='缩小座位图'
                onClick={() =>
                  setZoomLevel((z) => Math.max(0.5, +(z - 0.1).toFixed(2)))
                }
                className='size-7'
              >
                <ZoomOut className='size-3.5' />
              </Button>
              <span className='text-[10px] font-mono px-1 select-none text-muted-foreground min-w-8 text-center'>
                {Math.round(zoomLevel * 100)}%
              </span>
              <Button
                variant='ghost'
                size='icon'
                aria-label='放大座位图'
                onClick={() =>
                  setZoomLevel((z) => Math.min(2.0, +(z + 0.1).toFixed(2)))
                }
                className='size-7'
              >
                <ZoomIn className='size-3.5' />
              </Button>
              <Button
                variant='ghost'
                size='icon'
                aria-label='重置画布视图'
                title='重置缩放与位置'
                onClick={handleResetView}
                className='size-7 ml-0.5 text-muted-foreground hover:text-foreground'
              >
                <RotateCcw className='size-3' />
              </Button>
            </div>
          </div>
        </div>

        {/* 图例栏 */}
        <div className='flex flex-wrap items-center justify-between gap-2 text-[11px] text-muted-foreground pt-2 border-t border-dashed mt-2'>
          <div className='flex flex-wrap items-center gap-3.5'>
            <div className='flex items-center gap-1.5'>
              <span className='size-3 rounded border border-border bg-card' />
              <span>{t('legendAvailable')}</span>
            </div>
            <div className='flex items-center gap-1.5'>
              <span className='size-3 rounded bg-muted/60 border border-border/40' />
              <span>{t('legendOccupied')}</span>
            </div>
            <div className='flex items-center gap-1.5'>
              <span className='size-3 rounded bg-primary text-primary-foreground flex items-center justify-center font-bold text-[7px]'>
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
          <div className='flex items-center gap-1 text-[10px] text-muted-foreground/80 select-none'>
            <Move className='size-3' />
            <span>按住空白区域可自由拖动画布</span>
          </div>
        </div>
      </CardHeader>

      <CardContent className='space-y-3'>
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
                className='h-7 text-xs text-muted-foreground shadow-none'
              >
                <Trash2 className='size-3 mr-1' />
                {t('clearSelection')}
              </Button>
              <Button
                variant='outline'
                size='sm'
                onClick={onSaveFavorites}
                className='h-7 text-xs gap-1 border-dashed shadow-none'
              >
                <BookmarkCheck className='size-3 text-amber-500' />
                {t('saveToFavorites')}
              </Button>
              <Button
                variant='default'
                size='sm'
                onClick={handleApplyGrab}
                className='h-7 text-xs gap-1 shadow-none'
              >
                <Rocket className='size-3' />
                {t('applyGrab')}
              </Button>
            </div>
          </div>
        )}

        {/* 座位排布网格渲染容器 */}
        <div
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUpOrLeave}
          onMouseLeave={handleMouseUpOrLeave}
          className={`
            relative border border-dashed shadow-none rounded-lg bg-muted/15 p-4 overflow-hidden min-h-80 max-h-[580px] select-none
            ${isDragging ? 'cursor-grabbing' : 'cursor-grab'}
          `}
        >
          <div
            style={{
              transform: `translate(${panPosition.x}px, ${panPosition.y}px) scale(${zoomLevel})`,
              transformOrigin: '0 0',
              transition: isDragging ? 'none' : 'transform 0.1s ease-out',
            }}
            className='inline-block'
          >
            <div
              className='grid gap-1.5'
              style={{
                gridTemplateColumns: `repeat(${Math.max(totalCols, 1)}, minmax(36px, 36px))`,
              }}
            >
              {Array.from({ length: totalRows }, (_, rIdx) => {
                const row = rIdx + 1;
                return Array.from({ length: totalCols }, (_, cIdx) => {
                  const col = cIdx + 1;
                  const seat = seatsByCoord.get(`${col},${row}`);

                  if (!seat) {
                    // 空过道或空白占位
                    return (
                      <div
                        key={`empty-${col}-${row}`}
                        className='w-9 h-9 opacity-0 pointer-events-none'
                      />
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
