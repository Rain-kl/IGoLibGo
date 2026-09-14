// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { ArrowDown, ArrowUp, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import type { SeatRef } from '@/lib/services/igo/types';

interface SeatPreferenceListProps {
  seats: SeatRef[];
  onChange: (seats: SeatRef[]) => void;
  emptyText?: string;
}

export function SeatPreferenceList({
  seats,
  onChange,
  emptyText = '尚未添加候选座位',
}: SeatPreferenceListProps) {
  const moveUp = (index: number) => {
    if (index === 0) return;
    const next = [...seats];
    const temp = next[index - 1];
    next[index - 1] = next[index];
    next[index] = temp;
    onChange(next);
  };

  const moveDown = (index: number) => {
    if (index === seats.length - 1) return;
    const next = [...seats];
    const temp = next[index + 1];
    next[index + 1] = next[index];
    next[index] = temp;
    onChange(next);
  };

  const removeSeat = (index: number) => {
    onChange(seats.filter((_, i) => i !== index));
  };

  if (seats.length === 0) {
    return (
      <div className='p-6 text-center border border-dashed rounded-lg text-xs text-muted-foreground'>
        {emptyText}
      </div>
    );
  }

  return (
    <div className='space-y-1.5'>
      {seats.map((seat, index) => (
        <div
          key={seat.seat_key}
          className='flex items-center justify-between p-2 px-3 rounded-md border bg-card hover:bg-muted/30 transition-colors text-xs'
        >
          <div className='flex items-center gap-2'>
            <Badge
              variant={index === 0 ? 'default' : 'secondary'}
              className='size-5 p-0 flex items-center justify-center font-mono text-[10px]'
            >
              {index + 1}
            </Badge>
            <div>
              <span className='font-mono font-semibold text-foreground'>
                {seat.seat_name || seat.seat_key}
              </span>
              <span className='text-[10px] text-muted-foreground ml-2 font-mono'>
                Key: {seat.seat_key}
              </span>
            </div>
          </div>

          <div className='flex items-center gap-1'>
            <Button
              variant='ghost'
              size='icon'
              aria-label='上移优先级'
              disabled={index === 0}
              onClick={() => moveUp(index)}
              className='size-6'
            >
              <ArrowUp className='size-3' />
            </Button>
            <Button
              variant='ghost'
              size='icon'
              aria-label='下移优先级'
              disabled={index === seats.length - 1}
              onClick={() => moveDown(index)}
              className='size-6'
            >
              <ArrowDown className='size-3' />
            </Button>
            <Button
              variant='ghost'
              size='icon'
              aria-label='移除候选座位'
              onClick={() => removeSeat(index)}
              className='size-6 text-destructive hover:bg-destructive/10'
            >
              <Trash2 className='size-3' />
            </Button>
          </div>
        </div>
      ))}
    </div>
  );
}
