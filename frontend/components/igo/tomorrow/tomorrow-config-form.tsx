// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import Link from 'next/link';
import { useTranslations } from 'next-intl';
import { Building2, Clock, ExternalLink, MapPin } from 'lucide-react';
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
import type { BoundLibraryResponse, SeatRef } from '@/lib/services/igo/types';

interface TomorrowConfigFormProps {
  boundInfo: BoundLibraryResponse | null;
  seat: SeatRef | null;
  scheduledTime: string;
  onScheduledTimeChange: (val: string) => void;
}

export function TomorrowConfigForm({
  boundInfo,
  seat,
  scheduledTime,
  onScheduledTimeChange,
}: TomorrowConfigFormProps) {
  const t = useTranslations('igo.tomorrow');
  const bound = boundInfo?.library;

  return (
    <Card className='border-border/60 shadow-sm'>
      <CardHeader className='pb-3'>
        <div className='flex items-center gap-2'>
          <Clock className='size-4 text-primary' />
          <CardTitle className='text-base font-semibold'>
            明日预约目标与时间配置
          </CardTitle>
        </div>
        <CardDescription>
          设定次日场馆开放预约的准点时间（如前一天 22:00 或当天早晨 06:00）
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        <div className='grid grid-cols-1 sm:grid-cols-2 gap-4'>
          {/* 目标场馆 */}
          <div className='p-3 rounded-lg bg-muted/40 space-y-1 text-xs'>
            <div className='flex items-center justify-between'>
              <span className='text-muted-foreground flex items-center gap-1'>
                <Building2 className='size-3.5' />
                目标场馆
              </span>
              <Button
                variant='ghost'
                size='sm'
                asChild
                className='h-6 text-[11px] text-primary p-0'
              >
                <Link href='/venue'>切换场馆</Link>
              </Button>
            </div>
            <div className='font-semibold text-foreground text-sm'>
              {bound ? `${bound.name} · ${bound.floor}` : '未锁定场馆'}
            </div>
          </div>

          {/* 目标座位 */}
          <div className='p-3 rounded-lg bg-muted/40 space-y-1 text-xs'>
            <div className='flex items-center justify-between'>
              <span className='text-muted-foreground flex items-center gap-1'>
                <MapPin className='size-3.5' />
                {t('targetSeatTitle')}
              </span>
              <Button
                variant='ghost'
                size='sm'
                asChild
                className='h-6 text-[11px] text-primary p-0'
              >
                <Link href='/venue'>去选座</Link>
              </Button>
            </div>
            <div className='font-mono font-semibold text-foreground text-sm'>
              {seat
                ? seat.seat_name || seat.seat_key
                : '请先在【账户与场馆】中选中座位'}
            </div>
          </div>
        </div>

        <div className='space-y-1.5 max-w-sm'>
          <Label htmlFor='tomorrow-time' className='text-xs font-medium'>
            {t('scheduledTime')}
          </Label>
          <Input
            id='tomorrow-time'
            value={scheduledTime}
            onChange={(e) => onScheduledTimeChange(e.target.value)}
            placeholder={t('scheduledTimePlaceholder')}
            className='h-9 text-xs font-mono'
          />
          <p className='text-[10px] text-muted-foreground'>
            请输入标准时间字符串 (HH:mm:ss)，系统将在准点提前 10
            秒唤醒并发起预热请求
          </p>
        </div>
      </CardContent>
    </Card>
  );
}
