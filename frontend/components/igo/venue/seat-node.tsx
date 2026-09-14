// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { Bookmark, Tag } from 'lucide-react';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import type { SeatSnapshot } from '@/lib/services/igo/types';

interface SeatNodeProps {
  seat: SeatSnapshot;
  isSelected: boolean;
  selectionIndex?: number;
  label?: string;
  isFavorite?: boolean;
  onToggle: (seat: SeatSnapshot) => void;
}

export function SeatNode({
  seat,
  isSelected,
  selectionIndex,
  label,
  isFavorite,
  onToggle,
}: SeatNodeProps) {
  const isOccupied = seat.is_occupied;

  return (
    <TooltipProvider delayDuration={150}>
      <Tooltip>
        <TooltipTrigger asChild>
          <button
            type='button'
            disabled={isOccupied}
            onClick={() => onToggle(seat)}
            aria-label={`座位 ${seat.seat_name || seat.seat_key} ${isOccupied ? '已被占用' : '空闲可选'}`}
            className={`
              relative flex flex-col items-center justify-center rounded-md border text-xs font-mono transition-all duration-150 select-none
              w-10 h-10 p-0.5
              ${
                isSelected
                  ? 'bg-primary text-primary-foreground border-primary shadow-sm ring-2 ring-primary/30 scale-105 z-10'
                  : isOccupied
                    ? 'bg-muted/50 text-muted-foreground/50 border-border/40 cursor-not-allowed opacity-60'
                    : 'bg-card text-card-foreground border-border hover:border-primary/80 hover:bg-accent/40 cursor-pointer'
              }
            `}
          >
            {/* 序号或座位名 */}
            <span className='font-semibold text-[10px] leading-tight truncate max-w-full'>
              {seat.seat_name || seat.seat_key}
            </span>

            {/* 选中顺序角标 */}
            {isSelected && selectionIndex !== undefined && (
              <span className='absolute -top-1.5 -right-1.5 size-4 rounded-full bg-primary text-primary-foreground border border-background text-[9px] font-bold flex items-center justify-center shadow-xs'>
                {selectionIndex + 1}
              </span>
            )}

            {/* 收藏或标签小标记 */}
            <div className='flex items-center gap-0.5 mt-0.5'>
              {isFavorite && (
                <Bookmark className='size-2.5 fill-amber-500 text-amber-500' />
              )}
              {label && <Tag className='size-2.5 text-primary' />}
            </div>
          </button>
        </TooltipTrigger>
        <TooltipContent side='top' className='text-xs space-y-1 p-2'>
          <div className='font-semibold flex items-center gap-1.5'>
            <span>{seat.seat_name || seat.seat_key}</span>
            {isFavorite && (
              <span className='text-[10px] text-amber-500'>★ 已收藏</span>
            )}
          </div>
          <div className='text-muted-foreground text-[11px]'>
            状态：{isOccupied ? '已被占用' : '空闲可选'}
          </div>
          {label && (
            <div className='text-[11px] text-primary flex items-center gap-1'>
              <Tag className='size-3' />
              <span>备注：{label}</span>
            </div>
          )}
          <div className='text-[10px] text-muted-foreground/80 font-mono'>
            坐标: ({seat.x}, {seat.y}) · Key: {seat.seat_key}
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
