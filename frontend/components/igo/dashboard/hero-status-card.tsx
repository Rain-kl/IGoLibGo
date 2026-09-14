// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Activity, Clock, ShieldCheck } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import type { DashboardResponse } from '@/lib/services/igo/types';

interface HeroStatusCardProps {
  dashboard: DashboardResponse | null;
  userName?: string;
}

export function HeroStatusCard({
  dashboard,
  userName = '同学',
}: HeroStatusCardProps) {
  const t = useTranslations('igo.dashboard');

  const [currentTime, setCurrentTime] = React.useState<string>('--:--:--');
  const [currentDate, setCurrentDate] = React.useState<string>('--');

  React.useEffect(() => {
    const updateTime = () => {
      const now = new Date();
      setCurrentTime(now.toLocaleTimeString('zh-CN', { hour12: false }));
      setCurrentDate(
        now.toLocaleDateString('zh-CN', {
          month: 'long',
          day: 'numeric',
          weekday: 'long',
        }),
      );
    };
    updateTime();
    const timer = setInterval(updateTime, 1000);
    return () => clearInterval(timer);
  }, []);

  const greeting = React.useMemo(() => {
    const hour = new Date().getHours();
    if (hour < 12) return t('greetingMorning');
    if (hour < 18) return t('greetingAfternoon');
    return t('greetingEvening');
  }, [t]);

  const isAuthorized = dashboard?.authorized ?? false;
  const heroStatus =
    dashboard?.hero_status ||
    (isAuthorized ? t('heroReady') : t('heroWaitingAuth'));
  const heroDetail =
    dashboard?.hero_status_detail ||
    (isAuthorized ? t('heroReadyDesc') : t('heroWaitingAuthDesc'));

  return (
    <Card className='border-dashed shadow-none bg-gradient-to-r from-card via-card to-primary/5 overflow-hidden relative'>
      <CardContent className='p-6'>
        <div className='flex flex-col md:flex-row md:items-center md:justify-between gap-4'>
          <div className='space-y-1.5'>
            <div className='flex items-center gap-2'>
              <h2 className='text-xl font-semibold tracking-tight text-foreground'>
                {greeting}，{userName}
              </h2>
              {isAuthorized ? (
                <Badge
                  variant='outline'
                  className='gap-1 border-primary/40 text-primary'
                >
                  <ShieldCheck className='size-3' />
                  <span>已连接</span>
                </Badge>
              ) : (
                <Badge variant='secondary' className='text-muted-foreground'>
                  等待授权
                </Badge>
              )}
            </div>
            <p className='text-xs text-muted-foreground'>
              {t('greetingSubtitle')}
            </p>
            <div className='pt-2 flex items-center gap-2 text-xs'>
              <Activity className='size-3.5 text-primary' />
              <span className='font-semibold text-foreground'>
                {heroStatus}
              </span>
              <span className='text-muted-foreground'>· {heroDetail}</span>
            </div>
          </div>

          <div className='flex flex-col md:items-end justify-center font-mono space-y-0.5 border-t md:border-t-0 md:border-l border-border/40 pt-3 md:pt-0 md:pl-6 shrink-0'>
            <div className='text-2xl font-semibold tracking-wider text-foreground'>
              {currentTime}
            </div>
            <div className='text-xs text-muted-foreground flex items-center gap-1'>
              <Clock className='size-3' />
              <span>{currentDate}</span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
