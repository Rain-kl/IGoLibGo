// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import Link from 'next/link';
import { useTranslations } from 'next-intl';
import { Building2, ExternalLink, ListOrdered, Sliders } from 'lucide-react';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { SeatPreferenceList } from '@/components/igo/common/seat-preference-list';
import { GrabStrategyReminder } from './grab-strategy-reminder';
import type { BoundLibraryResponse, SeatRef } from '@/lib/services/igo/types';

interface GrabConfigFormProps {
  boundInfo: BoundLibraryResponse | null;
  seats: SeatRef[];
  onSeatsChange: (seats: SeatRef[]) => void;
  scheduledStart: string;
  onScheduledStartChange: (val: string) => void;
  strategy: string;
  onStrategyChange: (val: string) => void;
  minDelay: number;
  onMinDelayChange: (val: number) => void;
  maxDelay: number;
  onMaxDelayChange: (val: number) => void;
}

export function GrabConfigForm({
  boundInfo,
  seats,
  onSeatsChange,
  scheduledStart,
  onScheduledStartChange,
  strategy,
  onStrategyChange,
  minDelay,
  onMinDelayChange,
  maxDelay,
  onMaxDelayChange,
}: GrabConfigFormProps) {
  const t = useTranslations('igo.grab');
  const bound = boundInfo?.library;

  return (
    <div className='space-y-4'>
      {/* 1. 目标场馆卡片 */}
      <Card className='border-dashed shadow-none'>
        <CardHeader className='pb-3'>
          <div className='flex items-center justify-between'>
            <div className='flex items-center gap-2'>
              <Building2 className='size-4 text-primary' />
              <CardTitle className='text-base font-semibold'>
                {t('targetVenue')}
              </CardTitle>
            </div>
            <Button
              variant='ghost'
              size='sm'
              asChild
              className='h-7 text-xs text-primary gap-1 shadow-none'
            >
              <Link href='/venue'>
                <span>去切换场馆</span>
                <ExternalLink className='size-3' />
              </Link>
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {bound ? (
            <div className='p-3 rounded-lg bg-muted/40 flex items-center justify-between text-xs'>
              <div className='flex items-center gap-3'>
                <div className='size-8 rounded bg-primary/10 text-primary flex items-center justify-center font-semibold'>
                  {bound.name.charAt(0)}
                </div>
                <div>
                  <div className='font-semibold text-foreground text-sm'>
                    {bound.name}
                  </div>
                  <div className='text-muted-foreground text-[11px]'>
                    {bound.floor}
                  </div>
                </div>
              </div>
              <div className='text-right'>
                <div className='font-semibold text-primary'>
                  余 {bound.total_seats - bound.used_seats - bound.booked_seats}{' '}
                  座
                </div>
                <div className='text-[10px] text-muted-foreground'>
                  共 {bound.total_seats} 座
                </div>
              </div>
            </div>
          ) : (
            <div className='p-4 text-center border-dashed border rounded-lg text-xs text-muted-foreground'>
              {t('venueNotSelected')}
            </div>
          )}
        </CardContent>
      </Card>

      {/* 2. 优选座位优先级队列 */}
      <Card className='border-dashed shadow-none'>
        <CardHeader className='pb-3'>
          <div className='flex items-center justify-between'>
            <div className='flex items-center gap-2'>
              <ListOrdered className='size-4 text-primary' />
              <CardTitle className='text-base font-semibold'>
                {t('seatQueueTitle')} ({seats.length})
              </CardTitle>
            </div>
            <Button
              variant='outline'
              size='sm'
              asChild
              className='h-7 text-xs gap-1 border-dashed shadow-none'
            >
              <Link href='/venue'>
                <span>从座位图添加</span>
                <ExternalLink className='size-3' />
              </Link>
            </Button>
          </div>
          <CardDescription>{t('seatQueueDesc')}</CardDescription>
        </CardHeader>
        <CardContent className='space-y-3'>
          <SeatPreferenceList
            seats={seats}
            onChange={onSeatsChange}
            emptyText={t('emptyQueue')}
          />
          <GrabStrategyReminder seatCount={seats.length} />
        </CardContent>
      </Card>

      {/* 3. 策略与定时参数 */}
      <Card className='border-dashed shadow-none'>
        <CardHeader className='pb-3'>
          <div className='flex items-center gap-2'>
            <Sliders className='size-4 text-primary' />
            <CardTitle className='text-base font-semibold'>
              策略与定时参数
            </CardTitle>
          </div>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='grid grid-cols-1 sm:grid-cols-2 gap-4'>
            <div className='space-y-1.5'>
              <Label htmlFor='scheduled-time' className='text-xs font-medium'>
                {t('scheduledStart')}
              </Label>
              <Input
                id='scheduled-time'
                value={scheduledStart}
                onChange={(e) => onScheduledStartChange(e.target.value)}
                placeholder='留空即立即启动 (如 07:00:00)'
                className='h-8 text-xs font-mono shadow-none bg-background'
              />
              <p className='text-[10px] text-muted-foreground'>
                {t('scheduledStartHelp')}
              </p>
            </div>

            <div className='space-y-1.5'>
              <Label className='text-xs font-medium'>{t('strategyMode')}</Label>
              <Select value={strategy} onValueChange={onStrategyChange}>
                <SelectTrigger className='h-8 text-xs shadow-none bg-background'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='query_then_reserve' className='text-xs'>
                    {t('strategyQueryReserve')}
                  </SelectItem>
                  <SelectItem value='direct_reserve' className='text-xs'>
                    {t('strategyDirectReserve')}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className='grid grid-cols-2 gap-4 pt-1'>
            <div className='space-y-1.5'>
              <Label htmlFor='min-delay' className='text-xs font-medium'>
                {t('delayMin')}
              </Label>
              <Input
                id='min-delay'
                type='number'
                value={minDelay}
                onChange={(e) => onMinDelayChange(Number(e.target.value))}
                min={0}
                max={5000}
                className='h-8 text-xs font-mono shadow-none bg-background'
              />
            </div>
            <div className='space-y-1.5'>
              <Label htmlFor='max-delay' className='text-xs font-medium'>
                {t('delayMax')}
              </Label>
              <Input
                id='max-delay'
                type='number'
                value={maxDelay}
                onChange={(e) => onMaxDelayChange(Number(e.target.value))}
                min={0}
                max={5000}
                className='h-8 text-xs font-mono shadow-none bg-background'
              />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
