// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { BookOpen, Calendar, Clock, Info, ShieldCheck } from 'lucide-react';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import type { LibraryRuleResponse } from '@/lib/services/igo/types';

interface VenueRuleCardProps {
  rule: LibraryRuleResponse | null;
  loading: boolean;
}

export function VenueRuleCard({ rule, loading }: VenueRuleCardProps) {
  const t = useTranslations('igo.venue');

  if (!rule && !loading) {
    return null;
  }

  return (
    <Card className='border-dashed shadow-none'>
      <CardHeader className='pb-3'>
        <div className='flex items-center gap-2'>
          <BookOpen className='size-4 text-primary' />
          <CardTitle className='text-base font-semibold'>
            {t('ruleTitle')}
          </CardTitle>
        </div>
        <CardDescription>开放时间与预约保留规则详情</CardDescription>
      </CardHeader>
      <CardContent>
        {rule ? (
          <div className='grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs'>
            <div className='p-2.5 rounded-lg bg-muted/40 space-y-1'>
              <div className='flex items-center gap-1.5 text-muted-foreground'>
                <Clock className='size-3.5' />
                <span>开放时间</span>
              </div>
              <div className='font-medium text-foreground'>
                {rule.open_time_text || '--'} ~ {rule.close_time_text || '--'}
              </div>
            </div>

            <div className='p-2.5 rounded-lg bg-muted/40 space-y-1'>
              <div className='flex items-center gap-1.5 text-muted-foreground'>
                <Calendar className='size-3.5' />
                <span>提前预约</span>
              </div>
              <div className='font-medium text-foreground'>
                {rule.advance_booking
                  ? `${rule.advance_booking} 天前开放`
                  : '--'}
              </div>
            </div>

            <div className='p-2.5 rounded-lg bg-muted/40 space-y-1'>
              <div className='flex items-center gap-1.5 text-muted-foreground'>
                <Info className='size-3.5' />
                <span>座位保留时长</span>
              </div>
              <div className='font-medium text-foreground'>
                {rule.seat_ttl_minutes ? `${rule.seat_ttl_minutes} 分钟` : '--'}
              </div>
            </div>

            <div className='p-2.5 rounded-lg bg-muted/40 space-y-1'>
              <div className='flex items-center gap-1.5 text-muted-foreground'>
                <ShieldCheck className='size-3.5' />
                <span>暂离保留与续座</span>
              </div>
              <div className='font-medium text-foreground'>
                暂离 {rule.hold_ttl_minutes || '--'} 分钟 · 续座{' '}
                {rule.renew_time_minutes || '--'} 分钟
              </div>
            </div>
          </div>
        ) : (
          <div className='text-xs text-muted-foreground py-2'>
            正在加载场馆开放规则...
          </div>
        )}
      </CardContent>
    </Card>
  );
}
